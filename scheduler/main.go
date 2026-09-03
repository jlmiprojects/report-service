package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"blueassetgroup.com/reports-service/model"
	"blueassetgroup.com/reports-service/repository"
	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/client"
	"blueassetgroup.com/reports-service/shared/config"
	"github.com/nats-io/nats.go"
	"gopkg.in/gomail.v2"
)

var (
	repo *repository.MongoRepository
	nc   *nats.Conn
)

func main() {
	var err error

	// CONFIG IS READ FROM conf/application.json (viper)
	config, err := utils.NewConfig()

	if err != nil {
		panic(err)
	}

	if config.Logger == nil {
		config.Logger = &utils.LoggerConfig{Type: "json", Level: "DEBUG"}
	}

	if config.Nats == nil {
		config.Nats = &utils.NatsConfig{Uri: "nats://localhost:4222"}
	}

	nc, err = nats.Connect(config.Nats.Uri, nats.Name("report-client"), nats.Token(config.Nats.Token))

	if err != nil {
		panic(err)
	}

	utils.SetupLogging("0.0.0", config.Logger)

	slog.Info("Logger is initialized")

	repo, err = repository.NewMongo(config)

	if err != nil {
		panic(err)
	}

	t := time.Now().Format("15:04")

	schedules, err := repo.FinddReportScheduleByTime(t)

	if err != nil {
		panic(err)
	}

	slog.Info("Amount of schedules to run us ", "count", len(schedules))

	for _, schedule := range schedules {

		slog.Info("Schedule to process", "schedule", schedule)

		if schedule.RepeatWeekly && matchDay(schedule) {
			runReport(schedule, config.Email, "weekly")
		} else if schedule.RepeatMonthly {
			runReport(schedule, config.Email, "monthly")
		} else if schedule.RepeatDaily {
			runReport(schedule, config.Email, "daily")
		} else {
			continue
		}
	}

}
func sendEmail(name string, profileId string, downloadfile string, conf *config.EmailConfig) error {

	profile, err := client.GetProfileById(nc, profileId)

	if err != nil {
		slog.Error("Failed to find the profile. for id ", "profileid", profile)
		return err
	}

	slog.Info("Sending Email", "host", conf.Host,
		"port", conf.Port,
		"from", conf.From, "to", profile.Email, "subject", "Report for "+name, "message", "Plase find attached report")

	d := gomail.NewDialer(conf.Host, conf.Port,
		conf.UserName, conf.Password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	m := gomail.NewMessage()
	m.SetHeader("From", conf.From)
	m.SetHeader("To", profile.Email)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Report for "+name)
	m.SetBody("text/html", "Please find attached report")
	m.Attach(downloadfile)

	if err := d.DialAndSend(m); err != nil {
		slog.Error("Failed to send email", "error", err)
		return err
	}

	return nil
}
func runReport(schedule *model.ReportSchedule, conf *config.EmailConfig, _type string) error {
	slog.Info("Running report", "id", schedule.ID)

	report, err := repo.FindByID(schedule.ReportId.Hex())
	if err != nil {
		slog.Error("Failed to find report", "id", schedule.ReportId.Hex(), "error", err)
		return err
	}

	slog.Info("Report params is ", "params", report.Parameters)

	downloadfile := fmt.Sprintf("%s/%s.xlsx", conf.DownloadDir, report.Name)

	reportUrl := fmt.Sprintf("%s?id=%s&profile_id=%s&name=%s&type=csv", conf.ReportUrl, report.ID.Hex(), schedule.ProfileId.Hex(), report.Name)

	for k, v := range schedule.Parameters {

		if k == "start" && _type == "daily" {
			s := time.Now().AddDate(0, 0, -1).Format("2006-01-02") + " 00:00"
			v = url.QueryEscape(s)
		}
		if k == "end" && _type == "daily" {
			s := time.Now().AddDate(0, 0, -1).Format("2006-01-02") + " 23:59"
			v = url.QueryEscape(s)

		}

		if k == "start" && _type == "weekly" {
			s := time.Now().AddDate(0, 0, -7).Format("2006-01-02") + " 00:00"
			v = url.QueryEscape(s)

		}
		if k == "end" && _type == "weekly" {
			s := time.Now().AddDate(0, 0, -7).Format("2006-01-02") + " 23:59"
			v = url.QueryEscape(s)

		}

		if v == "start" && _type == "monthly" {
			s := time.Now().AddDate(0, -1, 0).Format("2006-01-02") + " 00:00"
			v = url.QueryEscape(s)

		}
		if k == "end" && _type == "monthly" {
			s := time.Now().AddDate(0, -1, 0).Format("2006-01-02") + " 23:59"
			v = url.QueryEscape(s)

		}

		reportUrl += fmt.Sprintf("&%s=%v", k, v)
	}

	//reportUrl = url.QueryEscape(reportUrl)

	err = downloadFile(downloadfile, reportUrl)
	if err != nil {
		slog.Error("Failed to download file", "error", err)
		return err
	}

	//"https://api.jlmiprojects.co.za/api/reports/run?&id=673ec5c7ddac41262273510c&name=contacts&profile_id=69c238bca84d8ebf1ed508fb&param_ts=1779641390459&type=csv")

	return sendEmail(report.Name, schedule.ProfileId.Hex(), downloadfile, conf)
}

func downloadFile(filepath string, url string) error {

	slog.Info("Downloading file", "url", url, "filepath", filepath)

	// 1. Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the server returned a success status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// 2. Create the local file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 3. Write the body to file (streaming)
	_, err = io.Copy(out, resp.Body)
	return err
}

func matchTime(schedule *model.ReportSchedule) bool {

	t, err := time.Parse("15:04", schedule.StartTime)
	if err != nil {
		slog.Error("Failed to parse start time for", "id", schedule.ID, "error", err)
		return false
	}

	if t.Hour() == time.Now().Hour() &&
		t.Minute() == time.Now().Minute() {
		return true
	}

	return false
}

func matchDay(schedule *model.ReportSchedule) bool {

	today := time.Now().Weekday()

	if schedule.RepeatDaily {
		if today == 0 && schedule.Sunday {
			return true
		} else if today == 1 && schedule.Monday {
			return true
		} else if today == 2 && schedule.Tuesday {
			return true
		} else if today == 3 && schedule.Wednesday {
			return true
		} else if today == 4 && schedule.Thursday {
			return true
		} else if today == 5 && schedule.Friday {
			return true
		} else if today == 6 && schedule.Saturday {
			return true
		} else {
			return false
		}

	} else {
		return false
	}
}
