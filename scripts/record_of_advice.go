//go:build ignore

// record_of_advice.go is a report script: real Go syntax interpreted by
// yaegi at request time (see handlers/scriptrunner.go), not compiled into
// this module's binary — the ignore tag above keeps `go build ./...`/`go vet
// ./...` from trying to build every script file as one Go package (they all
// declare `package script` and would otherwise collide).
//
// record_of_advice.go — Record of Advice for a completed broker-portal FNA.
// See conf/record_of_advice_report.json ("script": "record_of_advice").
// GET /run?name=record-of-advice&fna_id=<hex ObjectID> (add &type=pdf for the
// PDF variant). The report is "unavailable" until a quote is linked.
//
// This replaces the old "fna" mongo.aggregate data action + chained
// record_of_advice.js reshape with one script: the join pipeline runs here
// via ctx.Mongo(...).Aggregate, and the reshaping (label maps,
// currency/percent/date formatting, risk-tolerance bucketing, commission-line
// assembly) is ported line-for-line from record_of_advice.js into Go. The
// assembled map deliberately uses the JS script's original lowercase keys so
// reports/record_of_advice.html needs no changes.
//
// The label maps below are duplicated from broker-portal
// ui/models/fna/db.go (Category.Label / ProductCategory.Label /
// CoverType.Label) and ui/models/client/db.go — keep them in sync. The
// risk-tolerance bucket keys are likewise duplicated from broker-portal
// ui/models/riskquiz/risk_tolerance.json (scoring.buckets).
package script

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"blueassetgroup.com/reports-service/reportapi"
)

var months = []string{"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December"}

var categoryLabels = map[string]string{
	"LIFE_COVER":     "Life cover",
	"DISABILITY":     "Disability / income protection",
	"RETIREMENT":     "Retirement",
	"EDUCATION":      "Education",
	"EMERGENCY_FUND": "Emergency fund",
}

var productLabels = map[string]string{
	"LIFE_ASSURANCE":        "Life assurance",
	"DISABILITY_COVER":      "Disability cover",
	"INCOME_PROTECTION":     "Income protection",
	"RETIREMENT_ANNUITY":    "Retirement annuity",
	"ENDOWMENT":             "Endowment",
	"UNIT_TRUST":            "Unit trust",
	"MEDICAL_AID_GAP_COVER": "Medical aid gap cover",
	"OTHER":                 "Other",
}

var coverLabels = map[string]string{
	"LIFE":               "Life",
	"DISABILITY":         "Disability",
	"INCOME_PROTECTION":  "Income protection",
	"RETIREMENT_ANNUITY": "Retirement annuity",
	"MEDICAL_AID":        "Medical aid",
	"OTHER":              "Other",
}

// Quote enum labels — duplicated from broker-portal ui/models/quote/db.go
// (Type.Label / WithdrawType / CommissionKind) and ui/models/product/db.go
// (Family.Label). Keep in sync.
var quoteTypeLabels = map[string]string{"ONCE_OFF": "Once-off", "RECURRING": "Recurring"}
var withdrawTypeLabels = map[string]string{"BANK": "Bank", "NA": "Not applicable"}
var commissionKindLabels = map[string]string{
	"LOAN_ONCE_OFF":   "Loan · once-off",
	"DAILY_RECURRING": "Daily · recurring",
}
var productFamilyLabels = map[string]string{
	"STANDARD":        "Standard",
	"LEGACY":          "Legacy",
	"CAPITAL_PROTECT": "Capital Protect",
	"ZAR":             "ZAR",
}

// ---- generic helpers -------------------------------------------------------

func str(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func num(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	default:
		return 0
	}
}

func boolVal(v any) bool {
	b, _ := v.(bool)
	return b
}

func mapVal(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func sliceVal(v any) []any {
	s, _ := v.([]any)
	return s
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func joinNonEmpty(sep string, parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

func formatNumber(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// intOrDash mirrors the JS `v ? String(v) : "—"` pattern for numeric fields.
func intOrDash(v any) string {
	n := num(v)
	if n == 0 {
		return "—"
	}
	return strconv.Itoa(int(n))
}

// ---- formatting helpers (ported from record_of_advice.js) -----------------

// rand formats an integer-cents amount as "R 1 234 567" (no decimals).
func rand(cents any) string {
	v := math.Round(num(cents) / 100)
	neg := v < 0
	if neg {
		v = -v
	}
	s := strconv.FormatFloat(v, 'f', 0, 64)
	out := groupThousands(s)
	if neg {
		return "-R " + out
	}
	return "R " + out
}

func groupThousands(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var parts []string
	for n > 3 {
		parts = append([]string{s[n-3:]}, parts...)
		s = s[:n-3]
		n = len(s)
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, " ")
}

// pct turns a stored fraction (0.06) into a whole-number percent string ("6").
func pct(frac any) string {
	v := math.Round(num(frac)*100*100) / 100
	return formatNumber(v)
}

func titleCase(code string) string {
	if code == "" {
		return ""
	}
	words := strings.Split(strings.ToLower(code), "_")
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func label(m map[string]string, code string) string {
	if l, ok := m[code]; ok {
		return l
	}
	return titleCase(code)
}

// fmtDate accepts an RFC3339 string or "YYYY-MM-DD" and returns "2 January 2026".
func fmtDate(s string) string {
	if s == "" {
		return ""
	}
	if len(s) < 10 || s[4] != '-' || s[7] != '-' {
		return s
	}
	year, month, day := s[0:4], s[5:7], s[8:10]
	mi, err1 := strconv.Atoi(month)
	di, err2 := strconv.Atoi(day)
	if err1 != nil || err2 != nil || mi < 1 || mi > 12 {
		return s
	}
	return fmt.Sprintf("%d %s %s", di, months[mi-1], year)
}

func addressLines(a map[string]any) []string {
	lines := []string{}
	if a == nil {
		return lines
	}
	if l1 := str(a["line1"]); l1 != "" {
		lines = append(lines, l1)
	}
	if l2 := str(a["line2"]); l2 != "" {
		lines = append(lines, l2)
	}
	cityLine := joinNonEmpty(", ", str(a["city"]), str(a["state"]), str(a["postal_code"]))
	if cityLine != "" {
		lines = append(lines, cityLine)
	}
	if cc := str(a["country_code"]); cc != "" {
		lines = append(lines, cc)
	}
	return lines
}

func money(v any, monthly bool) string {
	if monthly {
		return rand(v) + " / month"
	}
	return rand(v)
}

func amount(v any) string { return rand(v) }

func row(labelText, value string) map[string]any {
	return map[string]any{"label": labelText, "value": value}
}

func Run(ctx reportapi.Context) (map[string]any, error) {
	fnaID := ctx.Query["fna_id"]

	res, err := ctx.Mongo("broker_portal").Aggregate("", "fna", []any{
		map[string]any{"$match": map[string]any{"_id": map[string]any{"$oid": fnaID}}},
		map[string]any{"$limit": 1},
		map[string]any{"$lookup": map[string]any{"from": "clients", "localField": "client_id", "foreignField": "_id", "as": "client"}},
		map[string]any{"$unwind": map[string]any{"path": "$client", "preserveNullAndEmptyArrays": true}},
		map[string]any{"$lookup": map[string]any{"from": "advisors", "localField": "advisor_id", "foreignField": "_id", "as": "advisor"}},
		map[string]any{"$unwind": map[string]any{"path": "$advisor", "preserveNullAndEmptyArrays": true}},
		map[string]any{"$lookup": map[string]any{"from": "brokerages", "localField": "advisor.brokerage_id", "foreignField": "_id", "as": "brokerage"}},
		map[string]any{"$unwind": map[string]any{"path": "$brokerage", "preserveNullAndEmptyArrays": true}},
		map[string]any{"$lookup": map[string]any{
			"from": "quotes",
			"let":  map[string]any{"fid": "$_id"},
			"pipeline": []any{
				map[string]any{"$match": map[string]any{"$expr": map[string]any{"$and": []any{
					map[string]any{"$eq": []any{"$fna_id", "$$fid"}},
					map[string]any{"$in": []any{"$status", []any{"SUBMITTED", "EMAILED", "ACCEPTED"}}},
				}}}},
				map[string]any{"$sort": map[string]any{"created_at": -1}},
				map[string]any{"$limit": 1},
			},
			"as": "quote",
		}},
		map[string]any{"$unwind": map[string]any{"path": "$quote", "preserveNullAndEmptyArrays": true}},
	})
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return map[string]any{"record_of_advice": map[string]any{"missing": true}}, nil
	}
	fna := res.Data[0]

	if fna["quote"] == nil {
		return map[string]any{"record_of_advice": map[string]any{
			"missing":        true,
			"missing_reason": "No quote has been generated for this financial needs analysis yet. The Record of Advice becomes available once the advisor has produced and submitted a quote for the client.",
		}}, nil
	}

	// ---- client ------------------------------------------------------

	c := mapVal(fna["client"])
	isCompany := str(c["type"]) == "COMPANY"
	clientName := str(c["company_name"])
	if !isCompany {
		clientName = joinNonEmpty(" ", str(c["name"]), str(c["surname"]))
	}
	idOrReg := str(c["id_number"])
	idOrRegLabel := "ID number"
	if isCompany {
		idOrReg = str(c["reg_no"])
		idOrRegLabel = "Registration number"
	}

	client := map[string]any{
		"is_company":        isCompany,
		"name":              firstNonEmpty(clientName, "—"),
		"code":              str(c["client_code"]),
		"id_or_reg":         idOrReg,
		"id_or_reg_label":   idOrRegLabel,
		"date_of_birth":     fmtDate(str(c["date_of_birth"])),
		"marital_status":    titleCase(str(c["marital_status"])),
		"employment_status": titleCase(str(c["employment_status"])),
		"occupation":        str(c["occupation"]),
		"email":             str(c["email"]),
		"mobile":            str(c["mobile"]),
		"address_lines":     addressLines(mapVal(c["address"])),
	}

	// ---- advisor -------------------------------------------------------

	av := mapVal(fna["advisor"])
	advisor := map[string]any{
		"name":   joinNonEmpty(" ", str(av["title"]), str(av["name"]), str(av["surname"])),
		"code":   str(av["advisor_code"]),
		"email":  str(av["email"]),
		"mobile": str(av["mobile"]),
	}

	// ---- brokerage (letterhead) --------------------------------------

	bk, hasBrokerage := fna["brokerage"].(map[string]any)
	brokerage := map[string]any{"present": hasBrokerage}
	if hasBrokerage {
		brokerage["name"] = str(bk["name"])
		brokerage["fsp_number"] = str(bk["fsp_number"])
		brokerage["reg_no"] = str(bk["reg_no"])
		brokerage["vat"] = str(bk["vat"])
		brokerage["contact_person"] = str(bk["contact_person"])
		brokerage["contact_email"] = str(bk["contact_email"])
		brokerage["contact_mobile"] = str(bk["contact_mobile"])
		brokerage["address_lines"] = addressLines(mapVal(bk["address"]))
	}

	// ---- fact-find ---------------------------------------------------

	ff := mapVal(fna["fact_find"])

	factfind := map[string]any{
		"income": []map[string]any{
			row("Gross monthly income", money(ff["gross_monthly_income"], false)),
			row("Net monthly income", money(ff["net_monthly_income"], false)),
		},
		"expenses": []map[string]any{
			row("Monthly living expenses", money(ff["monthly_living_expenses"], false)),
			row("Monthly debt repayments", money(ff["monthly_debt_repayments"], false)),
		},
		"assets": []map[string]any{
			row("Cash & savings", money(ff["cash_savings"], false)),
			row("Investments (unit trusts, shares)", money(ff["investments"], false)),
			row("Retirement fund — current value", money(ff["retirement_fund_value"], false)),
			row("Property value", money(ff["property_value"], false)),
			row("Monthly retirement contribution", money(ff["monthly_retirement_contribution"], false)),
		},
		"liabilities": []map[string]any{
			row("Home loan balance", money(ff["home_loan_balance"], false)),
			row("Vehicle finance balance", money(ff["vehicle_finance_balance"], false)),
			row("Other debt balance", money(ff["other_debt_balance"], false)),
		},
		"retirement_goal": []map[string]any{
			row("Target retirement age", intOrDash(ff["target_retirement_age"])),
			row("Target monthly retirement income (today's terms)", money(ff["target_monthly_retirement_income"], false)),
		},
		"other_goals":     str(ff["other_goals"]),
		"existing_cover":  []map[string]any{},
		"education_goals": []map[string]any{},
	}

	ec := sliceVal(ff["existing_cover"])
	existingCover := make([]map[string]any, 0, len(ec))
	for _, e := range ec {
		item := mapVal(e)
		existingCover = append(existingCover, map[string]any{
			"type":            label(coverLabels, str(item["type"])),
			"provider":        firstNonEmpty(str(item["provider"]), "—"),
			"cover_amount":    amount(item["cover_amount"]),
			"monthly_benefit": amount(item["monthly_benefit"]),
			"monthly_premium": amount(item["monthly_premium"]),
		})
	}
	factfind["existing_cover"] = existingCover

	eg := sliceVal(ff["education_goals"])
	educationGoals := make([]map[string]any, 0, len(eg))
	for _, e := range eg {
		item := mapVal(e)
		educationGoals = append(educationGoals, map[string]any{
			"dependant":  firstNonEmpty(str(item["dependant_name"]), "—"),
			"cost_today": amount(item["cost_today"]),
			"start_age":  intOrDash(item["start_age"]),
		})
	}
	factfind["education_goals"] = educationGoals

	// ---- needs analysis --------------------------------------------

	results := []map[string]any{}
	shortfalls := []map[string]any{}
	for _, r := range sliceVal(fna["results"]) {
		row := mapVal(r)
		gap := num(row["gap"])
		gapLabel := "Covered"
		if gap > 0 {
			gapLabel = "Shortfall"
		} else if gap < 0 {
			gapLabel = "Surplus"
		}
		monthly := boolVal(row["monthly"])
		item := map[string]any{
			"category":     label(categoryLabels, str(row["category"])),
			"need":         money(row["need"], monthly),
			"existing":     money(row["existing_provision"], monthly),
			"gap":          money(math.Abs(gap), monthly),
			"gap_label":    gapLabel,
			"is_shortfall": gap > 0,
			"is_surplus":   gap < 0,
			"monthly":      monthly,
			"note":         str(row["note"]),
		}
		results = append(results, item)
		if gap > 0 {
			shortfalls = append(shortfalls, item)
		}
	}

	// ---- recommendations -----------------------------------------

	recommendations := []map[string]any{}
	for _, r := range sliceVal(fna["recommendations"]) {
		rec := mapVal(r)
		recommendations = append(recommendations, map[string]any{
			"category": label(categoryLabels, str(rec["category"])),
			"product":  label(productLabels, str(rec["product"])),
			"notes":    str(rec["notes"]),
		})
	}

	// ---- risk tolerance ---------------------------------------------
	// From fna.questionnaire (riskquiz.Response, frozen at FNA completion).
	// Step 3 is skipped for non-SA clients, so questionnaire may be absent.

	var riskTolerance map[string]any
	if q, ok := fna["questionnaire"].(map[string]any); ok {
		bucketKey := str(q["result_bucket_key"])
		riskTolerance = map[string]any{
			"present":         true,
			"label":           str(q["result_label"]),
			"bucket_key":      bucketKey,
			"score":           strconv.FormatFloat(math.Round(num(q["score"])*10)/10, 'f', 1, 64),
			"method":          titleCase(str(q["method"])),
			"assessed_at":     fmtDate(str(q["submitted_at"])),
			"is_conservative": bucketKey == "conservative",
			"is_moderate":     bucketKey == "moderate",
			"is_aggressive":   bucketKey == "aggressive",
		}
	} else {
		riskTolerance = map[string]any{"present": false}
	}

	// ---- assumptions -------------------------------------------

	as := mapVal(fna["assumptions"])
	assumptions := map[string]any{
		"inflation_pct":              pct(as["inflation_rate"]),
		"growth_pct":                 pct(as["investment_growth_rate"]),
		"swr_pct":                    pct(as["safe_withdrawal_rate"]),
		"income_replacement_years":   intOrDash(as["income_replacement_years"]),
		"disability_replacement_pct": pct(as["disability_replacement_pct"]),
		"emergency_fund_months":      intOrDash(as["emergency_fund_months"]),
		"estate_and_funeral_costs":   amount(as["estate_and_funeral_costs"]),
	}

	// ---- quote (recommended investment) -----------------------------
	// fna.quote is the latest SUBMITTED quote linked to this FNA — guaranteed
	// present (see the early return above). All money fields are integer
	// cents; percentages (risk split, commission) are already whole numbers.

	qt := mapVal(fna["quote"])
	ps := mapVal(qt["product_snapshot"])
	qAssets := sliceVal(qt["assets"])
	qAsset := map[string]any{}
	if len(qAssets) > 0 {
		qAsset = mapVal(qAssets[0])
	}

	var qRisk map[string]any
	if risk, ok := qt["risk"].(map[string]any); ok {
		qRisk = map[string]any{
			"present":        true,
			"aggressive_pct": formatNumber(num(risk["aggressive_pct"])),
			"money_pct":      formatNumber(num(risk["money_pct"])),
			"locked":         boolVal(risk["locked"]),
		}
	} else {
		qRisk = map[string]any{"present": false}
	}

	var qLeverage map[string]any
	if lt, ok := qt["leverage_totals"].(map[string]any); ok {
		qLeverage = map[string]any{
			"present":            true,
			"total":              rand(lt["total"]),
			"in_money_fund":      rand(lt["in_money_fund"]),
			"in_aggressive_fund": rand(lt["in_aggressive_fund"]),
		}
	} else {
		qLeverage = map[string]any{"present": false}
	}

	wd := mapVal(qt["withdrawal"])
	wdType := str(wd["type"])
	wdDetail := ""
	if wdType != "" && wdType != "NA" {
		switch str(wd["mode"]) {
		case "SINGLE_PCT":
			wdDetail = formatNumber(num(wd["single_pct"])) + "% of the plan"
		case "TWO_BUCKET":
			wdDetail = fmt.Sprintf("Money fund %s%% · Aggressive fund %s%%",
				formatNumber(num(wd["money_fund_pct"])), formatNumber(num(wd["aggressive_fund_pct"])))
		}
	}

	commission := mapVal(qt["commission"])
	qCommissionLines := []map[string]any{}
	for _, l := range sliceVal(commission["lines"]) {
		line := mapVal(l)
		qCommissionLines = append(qCommissionLines, map[string]any{
			"kind":              label(commissionKindLabels, str(line["kind"])),
			"party":             firstNonEmpty(str(line["party_name"]), "—"),
			"recipient":         firstNonEmpty(str(line["recipient_name"]), "—"),
			"pct":               formatNumber(num(line["pct"])),
			"wallet_split_pct":  formatNumber(num(line["wallet_split_pct"])),
			"product_split_pct": formatNumber(num(line["product_split_pct"])),
			"wallet_amount":     rand(line["wallet_amount"]),
			"product_amount":    rand(line["product_amount"]),
		})
	}

	assetPct := num(qAsset["pct"])
	if assetPct == 0 {
		assetPct = 100
	}

	quote := map[string]any{
		"present":   true,
		"reference": str(qt["quote_code"]),
		"date":      fmtDate(str(qt["quote_date"])),
		"type":      label(quoteTypeLabels, str(qt["quote_type"])),
		"status":    str(qt["status"]),
		"plan_name": firstNonEmpty(str(ps["name"]), str(qt["product_code"]), "—"),
		"plan_code": str(qt["product_code"]),
		"family":    label(productFamilyLabels, str(ps["family"])),
		"amount":    rand(qt["amount"]),
		"asset": map[string]any{
			"fund_name": firstNonEmpty(str(qAsset["fund_name"]), str(qAsset["fund_code"]), "—"),
			"pct":       formatNumber(assetPct),
		},
		"risk":     qRisk,
		"leverage": qLeverage,
		"withdrawal": map[string]any{
			"type":   label(withdrawTypeLabels, firstNonEmpty(wdType, "NA")),
			"detail": wdDetail,
		},
		"commission": map[string]any{
			"currency":       firstNonEmpty(str(commission["currency"]), "ZAR"),
			"total_once_off": rand(commission["total_once_off"]),
			"lines":          qCommissionLines,
		},
		"notes": str(qt["notes"]),
	}

	// ---- assemble --------------------------------------------

	return map[string]any{"record_of_advice": map[string]any{
		"reference":       str(fna["fna_code"]),
		"status":          str(fna["status"]),
		"completed_at":    fmtDate(str(fna["completed_at"])),
		"client":          client,
		"advisor":         advisor,
		"brokerage":       brokerage,
		"factfind":        factfind,
		"results":         results,
		"shortfalls":      shortfalls,
		"has_shortfalls":  len(shortfalls) > 0,
		"recommendations": recommendations,
		"risk_tolerance":  riskTolerance,
		"quote":           quote,
		"assumptions":     assumptions,
	}}, nil
}
