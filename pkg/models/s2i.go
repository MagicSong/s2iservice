package models

import "time"

type S2IRequest struct {
	SourceURL            string `json:"source_url"`
	BuilderImage         string `json:"builder_image"`
	AddHost              string `json:"add_host,omitempty"`
	Tag                  string
	CallbackURL          string `json:"callback_url,omitempty"`
	ContextDir           string `json:"context_dir,omitempty"`
	RuntimeImage         string `json:"runtime_image,omitempty"`
	RuntimeArtifact      string `json:"runtime_artifact.omitempty"`
	EnvironmentVariables string `json:"environment_variables,omitempty"`
	ReuseMavenLocalRepo  bool   `json:"reuse_maven_local_repo,omitempty"`
	Export               bool   `json:"export,omitempty"`
	PushUsername         string `json:"push_username,omitempty"`
	PushPassword         string `json:"push_password"`
	Custom               string `json:"custom,omitempty"`
}
type S2IJob struct {
	ID           string `bson:"_id"`
	Username     string
	Parameters   []string  `json:"parameters,omitempty"`
	CreateTime   time.Time `json:"create_time,omitempty" bson:"create_time"`
	UpdateTime   time.Time `json:"update_time,omitempty" bson:"update_time"`
	Info         string    `json:"info,omitempty"`
	Status       JobStatus `json:"status,omitempty"`
	Retry        uint8     `json:"retry,omitempty"`
	ImageName    string    `json:"image_name" bson:"image_name"`
	Export       bool      `json:"export,omitempty"`
	PushUsername string    `json:"push_username,omitempty" bson:"push_username"`
	PushPassword string    `json:"push_password" bson:"push_password"`
}

type JobStatus string

const (
	Created    JobStatus = "Created"
	Processing JobStatus = "Processing"
	Error      JobStatus = "Error"
	Completed  JobStatus = "Completed"
	Terminated JobStatus = "Terminated"
)

type LogRow struct {
	Seq     int
	JobID   string    `bson:"builder_id"`
	Text    string    `bson:"log"`
	LogTime time.Time `bson:"create_time"`
	RetryID uint8     `bson:"retry_id"`
}

type FieldInfo struct {
	ID           uint `json:"id,omitempty" bson:"id"`
	TemplateID   uint `json:"template_id,omitempty" bson:"template_id"`
	Name         string
	TipsZH       string `json:"tips_zh,omitempty" bson:"tips_zh"`
	TipsEN       string `json:"tips_en,omitempty" bson:"tips_en"`
	FieldType    string `json:"field_type" bson:"field_type"`
	DefaultValue string `json:"default_value,omitempty" bson:"default_value"`
	Constraints  string `json:"constraints,omitempty"`
	Opts         string `json:"opts,omitempty"`
}
type S2ITemplate struct {
	ID            uint `json:"id,omitempty" bson:"id"`
	Name          string
	DescriptionEN string `json:"description_en,omitempty" bson:"description_en"`
	DescriptionZH string `json:"description_zh,omitempty" bson:"description_zh"`
	ICONPath      string `json:"icon_path,omitempty" bson:"icon_path"`
	BuilderImage  string `json:"builder_image" bson:"builder_image"`
	Status        string
	Language      string
	Used          int
	RuntimeImage  string      `json:"runtime_image,omitempty" bson:"runtime_image,omitempty"`
	Fields        []FieldInfo `json:"fields,omitempty"`
}
