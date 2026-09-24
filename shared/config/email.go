package config

type EmailConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	From        string `mapstructure:"from"`
	UserName    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	DownloadDir string `mapstructure:"download_dir"`
	ReportUrl   string `mapstructure:"report_url"`
}
