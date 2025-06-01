package config

type Config struct {
	Reqs struct {
		CreateRequestType         string `yaml:"create_req_type"`
		UpdateRequestType         string `yaml:"update_req_type"`
		DeleteQuestionRequestType string `yaml:"delete_question_req_type"`
		DeleteFormRequestType     string `yaml:"delete_form_req_type"`
	} `yaml:"reqs"`
	Urls struct {
		Redis    string `yaml:"redis"`
		Rabbitmq string `yaml:"rabbitmq"`
	} `yaml:"urls"`
	Exchange struct {
		Request string `yaml:"request"`
		Output  string `yaml:"output"`
	} `yaml:"exchange"`
	Queue struct {
		Request string `yaml:"request"`
		Output  string `yaml:"output"`
	} `yaml:"queue"`
	HealthCheck struct {
		Port string `yaml:"port"`
		Use  bool   `yaml:"use"`
	} `yaml:"health"`
}

func Init(path string) (*Config, error) {
	return &Config{
		Reqs: struct {
			CreateRequestType         string `yaml:"create_req_type"`
			UpdateRequestType         string `yaml:"update_req_type"`
			DeleteQuestionRequestType string `yaml:"delete_question_req_type"`
			DeleteFormRequestType     string `yaml:"delete_form_req_type"`
		}{
			CreateRequestType:         "request.form.created",
			UpdateRequestType:         "request.form.updated",
			DeleteQuestionRequestType: "request.question.deleted",
			DeleteFormRequestType:     "request.form.deleted",
		},
		Urls: struct {
			Redis    string `yaml:"redis"`
			Rabbitmq string `yaml:"rabbitmq"`
		}{
			Redis:    "redis:6379",
			Rabbitmq: "amqp://rabbitmq:5672",
		},
		Exchange: struct {
			Request string `yaml:"request"`
			Output  string `yaml:"output"`
		}{
			Request: "request",
			Output:  "output",
		},
		Queue: struct {
			Request string `yaml:"request"`
			Output  string `yaml:"output"`
		}{
			Request: "request",
			Output:  "output",
		},
	}, nil
}
