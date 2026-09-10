// record_of_advice.js — reshapes the joined FNA document fetched by the "fna"
// data action into a flat, presentation-ready `data.record_of_advice` object
// so reports/record_of_advice.html stays declarative.
//
// The label maps below are duplicated from broker-portal
// ui/models/fna/db.go (Category.Label / ProductCategory.Label /
// CoverType.Label) and ui/models/client/db.go — keep them in sync. The
// risk-tolerance bucket keys are likewise duplicated from broker-portal
// ui/models/riskquiz/risk_tolerance.json (scoring.buckets).

(function () {
    var rows = (data.fna && data.fna.data) ? data.fna.data : [];
    var fna = rows.length ? rows[0] : null;

    if (!fna) {
        data.record_of_advice = { missing: true };
        return;
    }

    // The Record of Advice is only available once the advisor has produced and
    // submitted a quote for this analysis. broker-portal links the latest
    // SUBMITTED quote to the FNA via quotes.fna_id; the "fna" data action joins
    // it as fna.quote. Until then the report is unavailable.
    if (!fna.quote) {
        data.record_of_advice = {
            missing: true,
            missing_reason: "No quote has been generated for this financial needs analysis yet. The Record of Advice becomes available once the advisor has produced and submitted a quote for the client."
        };
        return;
    }

    // ---- formatting helpers ------------------------------------------------

    function num(v) {
        var n = Number(v);
        return isNaN(n) ? 0 : n;
    }

    // rand formats an integer-cents amount as "R 1 234 567" (no decimals).
    function rand(cents) {
        var v = Math.round(num(cents) / 100);
        var neg = v < 0;
        v = Math.abs(v);
        var s = String(v);
        var out = "";
        while (s.length > 3) {
            out = " " + s.slice(s.length - 3) + out;
            s = s.slice(0, s.length - 3);
        }
        out = s + out;
        return (neg ? "-R " : "R ") + out;
    }

    // pct turns a stored fraction (0.06) into a whole-number percent string ("6").
    function pct(frac) {
        var v = Math.round(num(frac) * 100 * 100) / 100;
        return String(v);
    }

    function titleCase(code) {
        if (!code) return "";
        return String(code).toLowerCase().split("_").map(function (w) {
            return w ? w.charAt(0).toUpperCase() + w.slice(1) : w;
        }).join(" ");
    }

    var MONTHS = ["January", "February", "March", "April", "May", "June",
        "July", "August", "September", "October", "November", "December"];

    // fmtDate accepts an RFC3339 string or "YYYY-MM-DD" and returns "2 January 2026".
    function fmtDate(s) {
        if (!s) return "";
        var m = String(s).match(/^(\d{4})-(\d{2})-(\d{2})/);
        if (!m) return String(s);
        var d = parseInt(m[3], 10);
        return d + " " + MONTHS[parseInt(m[2], 10) - 1] + " " + m[1];
    }

    var CATEGORY_LABELS = {
        LIFE_COVER: "Life cover",
        DISABILITY: "Disability / income protection",
        RETIREMENT: "Retirement",
        EDUCATION: "Education",
        EMERGENCY_FUND: "Emergency fund"
    };
    var PRODUCT_LABELS = {
        LIFE_ASSURANCE: "Life assurance",
        DISABILITY_COVER: "Disability cover",
        INCOME_PROTECTION: "Income protection",
        RETIREMENT_ANNUITY: "Retirement annuity",
        ENDOWMENT: "Endowment",
        UNIT_TRUST: "Unit trust",
        MEDICAL_AID_GAP_COVER: "Medical aid gap cover",
        OTHER: "Other"
    };
    var COVER_LABELS = {
        LIFE: "Life",
        DISABILITY: "Disability",
        INCOME_PROTECTION: "Income protection",
        RETIREMENT_ANNUITY: "Retirement annuity",
        MEDICAL_AID: "Medical aid",
        OTHER: "Other"
    };

    // Quote enum labels — duplicated from broker-portal ui/models/quote/db.go
    // (Type.Label / WithdrawType / CommissionKind) and ui/models/product/db.go
    // (Family.Label). Keep in sync.
    var QUOTE_TYPE_LABELS = { ONCE_OFF: "Once-off", RECURRING: "Recurring" };
    var WITHDRAW_TYPE_LABELS = { BANK: "Bank", NA: "Not applicable" };
    var COMMISSION_KIND_LABELS = {
        LOAN_ONCE_OFF: "Loan · once-off",
        DAILY_RECURRING: "Daily · recurring"
    };
    var PRODUCT_FAMILY_LABELS = {
        STANDARD: "Standard",
        LEGACY: "Legacy",
        CAPITAL_PROTECT: "Capital Protect",
        ZAR: "ZAR"
    };

    function label(map, code) {
        return map[code] || titleCase(code);
    }

    function addressLines(a) {
        if (!a) return [];
        var lines = [];
        if (a.line1) lines.push(a.line1);
        if (a.line2) lines.push(a.line2);
        var cityLine = [a.city, a.state, a.postal_code].filter(function (x) { return !!x; }).join(", ");
        if (cityLine) lines.push(cityLine);
        if (a.country_code) lines.push(a.country_code);
        return lines;
    }

    function money(v, monthly) {
        return rand(v) + (monthly ? " / month" : "");
    }

    function amount(v) { return rand(v); }

    // ---- client ----------------------------------------------------------

    var c = fna.client || {};
    var isCompany = c.type === "COMPANY";
    var clientName = isCompany
        ? (c.company_name || "")
        : [c.name, c.surname].filter(function (x) { return !!x; }).join(" ");

    var client = {
        is_company: isCompany,
        name: clientName || "—",
        code: c.client_code || "",
        id_or_reg: isCompany ? (c.reg_no || "") : (c.id_number || ""),
        id_or_reg_label: isCompany ? "Registration number" : "ID number",
        date_of_birth: fmtDate(c.date_of_birth),
        marital_status: titleCase(c.marital_status),
        employment_status: titleCase(c.employment_status),
        occupation: c.occupation || "",
        email: c.email || "",
        mobile: c.mobile || "",
        address_lines: addressLines(c.address)
    };

    // ---- advisor -------------------------------------------------------

    var av = fna.advisor || {};
    var advisor = {
        name: [av.title, av.name, av.surname].filter(function (x) { return !!x; }).join(" "),
        code: av.advisor_code || "",
        email: av.email || "",
        mobile: av.mobile || ""
    };

    // ---- brokerage (letterhead) --------------------------------------

    var bk = fna.brokerage || {};
    var brokerage = {
        present: !!fna.brokerage,
        name: bk.name || "",
        fsp_number: bk.fsp_number || "",
        reg_no: bk.reg_no || "",
        vat: bk.vat || "",
        contact_person: bk.contact_person || "",
        contact_email: bk.contact_email || "",
        contact_mobile: bk.contact_mobile || "",
        address_lines: addressLines(bk.address)
    };

    // ---- fact-find ---------------------------------------------------

    var ff = fna.fact_find || {};

    function row(labelText, value) { return { label: labelText, value: value }; }

    var factfind = {
        income: [
            row("Gross monthly income", money(ff.gross_monthly_income)),
            row("Net monthly income", money(ff.net_monthly_income))
        ],
        expenses: [
            row("Monthly living expenses", money(ff.monthly_living_expenses)),
            row("Monthly debt repayments", money(ff.monthly_debt_repayments))
        ],
        assets: [
            row("Cash & savings", money(ff.cash_savings)),
            row("Investments (unit trusts, shares)", money(ff.investments)),
            row("Retirement fund — current value", money(ff.retirement_fund_value)),
            row("Property value", money(ff.property_value)),
            row("Monthly retirement contribution", money(ff.monthly_retirement_contribution))
        ],
        liabilities: [
            row("Home loan balance", money(ff.home_loan_balance)),
            row("Vehicle finance balance", money(ff.vehicle_finance_balance)),
            row("Other debt balance", money(ff.other_debt_balance))
        ],
        retirement_goal: [
            row("Target retirement age", ff.target_retirement_age ? String(ff.target_retirement_age) : "—"),
            row("Target monthly retirement income (today's terms)", money(ff.target_monthly_retirement_income))
        ],
        other_goals: ff.other_goals || "",
        existing_cover: [],
        education_goals: []
    };

    var ec = ff.existing_cover || [];
    for (var i = 0; i < ec.length; i++) {
        factfind.existing_cover.push({
            type: label(COVER_LABELS, ec[i].type),
            provider: ec[i].provider || "—",
            cover_amount: amount(ec[i].cover_amount),
            monthly_benefit: amount(ec[i].monthly_benefit),
            monthly_premium: amount(ec[i].monthly_premium)
        });
    }

    var eg = ff.education_goals || [];
    for (var j = 0; j < eg.length; j++) {
        factfind.education_goals.push({
            dependant: eg[j].dependant_name || "—",
            cost_today: amount(eg[j].cost_today),
            start_age: eg[j].start_age ? String(eg[j].start_age) : "—"
        });
    }

    // ---- needs analysis --------------------------------------------

    var results = [];
    var shortfalls = [];
    var rs = fna.results || [];
    for (var k = 0; k < rs.length; k++) {
        var gap = num(rs[k].gap);
        var gapLabel = gap > 0 ? "Shortfall" : (gap < 0 ? "Surplus" : "Covered");
        var item = {
            category: label(CATEGORY_LABELS, rs[k].category),
            need: money(rs[k].need, rs[k].monthly),
            existing: money(rs[k].existing_provision, rs[k].monthly),
            gap: money(Math.abs(gap), rs[k].monthly),
            gap_label: gapLabel,
            is_shortfall: gap > 0,
            is_surplus: gap < 0,
            monthly: !!rs[k].monthly,
            note: rs[k].note || ""
        };
        results.push(item);
        if (gap > 0) shortfalls.push(item);
    }

    // ---- recommendations -----------------------------------------

    var recommendations = [];
    var recs = fna.recommendations || [];
    for (var r = 0; r < recs.length; r++) {
        recommendations.push({
            category: label(CATEGORY_LABELS, recs[r].category),
            product: label(PRODUCT_LABELS, recs[r].product),
            notes: recs[r].notes || ""
        });
    }

    // ---- risk tolerance ---------------------------------------------
    // From fna.questionnaire (riskquiz.Response, frozen at FNA completion).
    // Step 3 is skipped for non-SA clients, so questionnaire may be absent.
    var q = fna.questionnaire;
    var riskTolerance = q
        ? {
            present: true,
            label: q.result_label || "",
            bucket_key: q.result_bucket_key || "",
            // score is a raw average of 0..2 option scores; show 1 decimal, no
            // denominator (buckets are not snapshotted on the Response).
            score: (Math.round(num(q.score) * 10) / 10).toFixed(1),
            method: titleCase(q.method || ""),           // "Average"
            assessed_at: fmtDate(q.submitted_at),
            // scale markers — keys duplicated from broker-portal
            // ui/models/riskquiz/risk_tolerance.json scoring.buckets; keep in sync.
            is_conservative: q.result_bucket_key === "conservative",
            is_moderate: q.result_bucket_key === "moderate",
            is_aggressive: q.result_bucket_key === "aggressive"
        }
        : { present: false };

    // ---- assumptions -------------------------------------------

    var as = fna.assumptions || {};
    var assumptions = {
        inflation_pct: pct(as.inflation_rate),
        growth_pct: pct(as.investment_growth_rate),
        swr_pct: pct(as.safe_withdrawal_rate),
        income_replacement_years: as.income_replacement_years ? String(as.income_replacement_years) : "—",
        disability_replacement_pct: pct(as.disability_replacement_pct),
        emergency_fund_months: as.emergency_fund_months ? String(as.emergency_fund_months) : "—",
        estate_and_funeral_costs: amount(as.estate_and_funeral_costs)
    };

    // ---- quote (recommended investment) -----------------------------
    // fna.quote is the latest SUBMITTED quote linked to this FNA — guaranteed
    // present (see the early return above). All money fields are integer cents;
    // percentages (risk split, commission) are already whole numbers.

    var qt = fna.quote;
    var ps = qt.product_snapshot || {};
    var qAsset = (qt.assets && qt.assets.length) ? qt.assets[0] : {};

    var qRisk = qt.risk
        ? {
            present: true,
            aggressive_pct: String(num(qt.risk.aggressive_pct)),
            money_pct: String(num(qt.risk.money_pct)),
            locked: !!qt.risk.locked
        }
        : { present: false };

    var lt = qt.leverage_totals;
    var qLeverage = lt
        ? {
            present: true,
            total: rand(lt.total),
            in_money_fund: rand(lt.in_money_fund),
            in_aggressive_fund: rand(lt.in_aggressive_fund)
        }
        : { present: false };

    var wd = qt.withdrawal || {};
    var wdDetail = "";
    if (wd.type && wd.type !== "NA") {
        if (wd.mode === "SINGLE_PCT") {
            wdDetail = num(wd.single_pct) + "% of the plan";
        } else if (wd.mode === "TWO_BUCKET") {
            wdDetail = "Money fund " + num(wd.money_fund_pct) + "% · Aggressive fund " + num(wd.aggressive_fund_pct) + "%";
        }
    }

    var qCommissionLines = [];
    var qcl = (qt.commission && qt.commission.lines) ? qt.commission.lines : [];
    for (var ci = 0; ci < qcl.length; ci++) {
        qCommissionLines.push({
            kind: label(COMMISSION_KIND_LABELS, qcl[ci].kind),
            party: qcl[ci].party_name || "—",
            recipient: qcl[ci].recipient_name || "—",
            pct: String(num(qcl[ci].pct)),
            wallet_split_pct: String(num(qcl[ci].wallet_split_pct)),
            product_split_pct: String(num(qcl[ci].product_split_pct)),
            wallet_amount: rand(qcl[ci].wallet_amount),
            product_amount: rand(qcl[ci].product_amount)
        });
    }

    var quote = {
        present: true,
        reference: qt.quote_code || "",
        date: fmtDate(qt.quote_date),
        type: label(QUOTE_TYPE_LABELS, qt.quote_type),
        status: qt.status || "",
        plan_name: ps.name || qt.product_code || "—",
        plan_code: qt.product_code || "",
        family: label(PRODUCT_FAMILY_LABELS, ps.family),
        amount: rand(qt.amount),
        asset: {
            fund_name: qAsset.fund_name || qAsset.fund_code || "—",
            pct: String(num(qAsset.pct || 100))
        },
        risk: qRisk,
        leverage: qLeverage,
        withdrawal: {
            type: label(WITHDRAW_TYPE_LABELS, wd.type || "NA"),
            detail: wdDetail
        },
        commission: {
            currency: (qt.commission && qt.commission.currency) || "ZAR",
            total_once_off: rand(qt.commission ? qt.commission.total_once_off : 0),
            lines: qCommissionLines
        },
        notes: qt.notes || ""
    };

    // ---- assemble --------------------------------------------

    data.record_of_advice = {
        reference: fna.fna_code || "",
        status: fna.status || "",
        completed_at: fmtDate(fna.completed_at),
        client: client,
        advisor: advisor,
        brokerage: brokerage,
        factfind: factfind,
        results: results,
        shortfalls: shortfalls,
        has_shortfalls: shortfalls.length > 0,
        recommendations: recommendations,
        risk_tolerance: riskTolerance,
        quote: quote,
        assumptions: assumptions
    };
})();
