# Source2Image 设计参考

为了降低负担，个人建议参考Openshift的设计，地址在https://139.198.120.203:8443/console/catalog，用户名和密码都是system，首页提供几个现成模板，用户点击语言和框架，然后配置一下仓库地址等参数，后台运行完成之后告知用户。现在需要输入的参数有：Source URL（git地址），Tag（镜像推送位置），Env环境变量。除此之外，还有针对特定语言的一些参数，现在还没有完整的测试。

如果来得及可以设计用户自定义Source2Image的页面，那么需要的参数如下：
```go
type S2IRequest struct {
	SourceURL            string `json:"source_url"`//git地址
	BuilderImage         string `json:"builder_image"`//最终镜像的基础镜像或者用于编译的镜像（如果指定了runtime image）。
	Tag                  string //镜像推送地址
	CallbackURL          string `json:"callback_url,omitempty"`//完成之后执行的URL
	ContextDir           string `json:"context_dir,omitempty"`//代码的位置，相对路径
	RuntimeImage         string `json:"runtime_image,omitempty"`//运行时镜像（默认上面的BuilderImage为运行时镜像，如果单独指定此项的话，表示build和运行是分开的）
	RuntimeArtifact      string `json:"runtime_artifact.omitempty"`//编译文件的位置（和runtimeImage字段需要同时指定）
	EnvironmentVariables string `json:"environment_variables,omitempty"`//环境变量
	Custom               string `json:"custom,omitempty"`//用户自己编辑s2i参数，如果不为空，覆盖上面所有
}
```
## 日志
### request
URL:    s2i.kubesphere.com/api/v1alpha1/jobs/:jid/logs?from_id=0&retry_id=0
其中的jid是任务id，from_id是日志的起始id，默认从0开始，如果需要实时日志的话，这个from_id需要增加。retry_id表明任务运行重试ID,默认为0，非0表示该任务第几次重试时的日志。
### response
```json
[
    {
        "JobID": 1,
        "Text": "tar: src/.gitignore: time stamp 2018-09-12 04:15:44 is 1.470437366 s in the future",
        "LogTime": "2018-09-12T04:15:46Z",
        "RetryID": 0
    },
    {
        "JobID": 1,
        "Text": "tar: src/README.md: time stamp 2018-09-12 04:15:44 is 1.445028489 s in the future",
        "LogTime": "2018-09-12T04:15:46Z",
        "RetryID": 0
    },
    {
        "JobID": 1,
        "Text": "tar: src/conf/reload.py: time stamp 2018-09-12 04:15:44 is 1.444861131 s in the future",
        "LogTime": "2018-09-12T04:15:46Z",
        "RetryID": 0
    },
    {
        "JobID": 1,
        "Text": "tar: src/conf: time stamp 2018-09-12 04:15:44 is 1.444834403 s in the future",
        "LogTime": "2018-09-12T04:15:46Z",
        "RetryID": 0
    }
]
```

## 模板
> 目前alpha版本只支持基于模板的Source2Image功能，后端提供一系列模板供前端使用。

### request

GET  http://s2i.kubesphere.com/api/v1alpha1/templates HTTP/1.1
Authorization: Basic admin admins

### response
```json
[
  {
    "id": 6,
    "Name": "Java+Maven",
    "description_en": "Build Java source code into docker image with Maven (openjdk 1.8.0)",
    "description_zh": "将Java代码一步制成Docker镜像（maven项目,openjdk 1.8.0）",
    "builder_image": "harbor.kubesphere.com/devops/s2i-maven-java:latest",
    "Status": "beta",
    "Language": "java",
    "Used": 0,
    "fields": [
      {
        "id": 1, 
        "template_id": 6,
        "Name": "Source",
        "tips_zh": "代码仓库地址，目前支持git和http/https开头的仓库地址",
        "tips_en": "url of source code, currently only git and http/https are supported",
        "field_type": "text",
        "constraints": "URL"
      },
      {
        "id": 2,
        "template_id": 6,
        "Name": "MVN_GOALS",
        "tips_zh": "指定maven执行的命令",
        "tips_en": "Maven goals to execute",
        "field_type": "text,environment",
        "default_value": "clean package"
      },
      {
        "id": 3,
        "template_id": 6,
        "Name": "MVN_SKIP_TESTS",
        "tips_zh": "指定是否跳过测试",
        "tips_en": "If tests should be skipped",
        "field_type": "checkbox,environment",
        "default_value": "true"
      }
    ]
  },
  {
    "id": 7,
    "Name": "Java+Maven+Tomcat",
    "description_en": "Platform for building and running Java applications on Apache-Tomcat",
    "description_zh": "将Tomcat程序一步制成Docker镜像",
    "builder_image": "sarcouy/s2i-tomcat",
    "Status": "beta",
    "Language": "java",
    "Used": 0,
    "fields": [
      {
        "id": 4,
        "template_id": 7,
        "Name": "Source",
        "tips_zh": "代码仓库地址，目前支持git和http/https开头的仓库地址",
        "tips_en": "url of source code, currently only git and http/https are supported",
        "field_type": "text",
        "constraints": "URL"
      },
      {
        "id": 5,
        "template_id": 7,
        "Name": "MVN_ARGS",
        "tips_zh": "指定maven执行的命令",
        "tips_en": "Maven goals to execute",
        "field_type": "text,environment",
        "default_value": "clean package -DskipTests"
      },
      {
        "id": 6,
        "template_id": 7,
        "Name": "Image_Version",
        "tips_zh": "镜像版本",
        "tips_en": "Image versions",
        "field_type": "select,tag",
        "default_value": "8-jdk8-mvn3.5.0",
        "opts": "6-jdk7-mvn3.2.5,6-jdk7-mvn3.3.9,6-jdk8-mvn3.2.5,6-jdk8-mvn3.3.9,7-jdk7-mvn3.2.5,7-jdk7-mvn3.3.9,7-jdk8-mvn3.2.5,7-jdk8-mvn3.3.9,8-jdk7-mvn3.2.5,8-jdk7-mvn3.3.9,8-jdk8-mvn3.2.5,8-jdk8-mvn3.3.9,8-jdk7-mvn3.5.0,8-jdk8-mvn3.5.0,8.5-jdk8-mvn3.2.5,8.5-jdk8-mvn3.3.9,8.5-jdk8-mvn3.5.0"
      },
      {
        "id": 7,
        "template_id": 7,
        "Name": "WAR_NAME",
        "tips_zh": "war包的名称，包括后缀.war",
        "tips_en": "Name of the war file to move into webapps directory after maven build WAR_NAME=myApp.war",
        "field_type": "text,environment"
      }
    ]
  },
  {
    "id": 8,
    "Name": "python3.5 + Django",
    "description_en": "Django is a high-level Python Web framework that encourages rapid development and clean, pragmatic design",
    "description_zh": "将Python web程序一步制成Docker镜像,Django框架+Python3.5",
    "builder_image": "centos/python-35-centos7",
    "Status": "beta",
    "Language": "python",
    "Used": 0,
    "fields": [
      {
        "id": 8,
        "template_id": 8,
        "Name": "Source",
        "tips_zh": "代码仓库地址，目前支持git和http/https开头的仓库地址",
        "tips_en": "url of source code, currently only git and http/https are supported",
        "field_type": "text",
        "constraints": "URL"
      },
      {
        "id": 9,
        "template_id": 8,
        "Name": "APP_SCRIPT",
        "tips_zh": "APP启动的shell脚本",
        "tips_en": "Used to run the application from a script file",
        "field_type": "text,environment",
        "default_value": "app.sh"
      },
      {
        "id": 10,
        "template_id": 8,
        "Name": "APP_FILE",
        "tips_zh": "APP启动的python脚本",
        "tips_en": "Used to run the application from a Python script",
        "field_type": "text,environment",
        "default_value": "app.py"
      },
      {
        "id": 11,
        "template_id": 8,
        "Name": "APP_HOME",
        "tips_zh": "APP启动根目录",
        "tips_en": "specify a sub-directory in which the application to be run is contained",
        "field_type": "text,environment,optional"
      }
    ]
  },
  {
    "id": 9,
    "Name": "Nodejs",
    "description_en": "Node.js is an open-source, cross-platform JavaScript run-time environment that executes JavaScript code outside of a browser",
    "description_zh": "将Nodejs程序打包成镜像，然后部署到各个地方",
    "builder_image": "fortinj66/centos7-s2i-nodejs:8.9.4",
    "Status": "beta",
    "Language": "javascript",
    "Used": 0,
    "fields": [
      {
        "id": 12,
        "template_id": 9,
        "Name": "Source",
        "tips_zh": "代码仓库地址，目前支持git和http/https开头的仓库地址",
        "tips_en": "url of source code, currently only git and http/https are supported",
        "field_type": "text",
        "constraints": "URL"
      }
    ]
  },
  {
    "id": 11,
    "Name": "php 5.6",
    "description_en": "PHP is a widely-used open source general-purpose scripting language that is especially suited for web development and can be embedded into HTML.",
    "description_zh": "将PHP web程序打包成镜像，然后部署到各个地方",
    "builder_image": "rvhoyt/s2i-php-56",
    "Status": "beta",
    "Language": "php",
    "Used": 0,
    "fields": [
      {
        "id": 18,
        "template_id": 11,
        "Name": "Source",
        "tips_zh": "代码仓库地址，目前支持git和http/https开头的仓库地址",
        "tips_en": "url of source code, currently only git and http/https are supported",
        "field_type": "text",
        "constraints": "URL"
      },
      {
        "id": 19,
        "template_id": 11,
        "Name": "DISPLAY_ERRORS",
        "tips_zh": "控制PHP是否会输出错误信息",
        "tips_en": "Controls whether or not and where PHP will output errors, notices and warnings",
        "field_type": "select,environment",
        "default_value": "on",
        "opts": "on,off"
      },
      {
        "id": 20,
        "template_id": 11,
        "Name": "SESSION_NAME",
        "tips_zh": "session字段的名称",
        "tips_en": "name of session",
        "field_type": "text,environment",
        "default_value": "PHPSESSID"
      },
      {
        "id": 21,
        "template_id": 11,
        "Name": "SESSION_HANDLER",
        "tips_zh": "session的管理方式",
        "tips_en": "Method for saving sessions",
        "field_type": "text,environment",
        "default_value": "files"
      },
      {
        "id": 22,
        "template_id": 11,
        "Name": "DOCUMENTROOT",
        "tips_zh": "应用的根目录",
        "tips_en": "Path that defines the DocumentRoot for your application",
        "field_type": "text,environment",
        "default_value": "/"
      }
    ]
  }
]
```
返回的是一个Template的数组。Template对象有几个重要的属性，一个是`builder_image`，这个属性前端在生成s2ijob的时候要用到，目前这个属性是readonly的，不能让用户修改；还有一个就是`fields`属性，`fields`告诉前端应该如何生成这个模板的页面。fields是一个`fieldinfo`的数组，`fieldinfo`的定义如下(下面是`go`语言的定义，具体json中的字段参看上面的reponse)：
```go
type FieldInfo struct {
	ID           uint `json:"id,omitempty" db:"id"`
	TemplateID   uint `json:"template_id,omitempty" db:"template_id"`
	Name         string
	TipsZH       string `json:"tips_zh,omitempty" db:"tips_zh"`
	TipsEN       string `json:"tips_en,omitempty" db:"tips_en"`
	FieldType    string `json:"field_type" db:"field_type"`
	DefaultValue string `json:"default_value,omitempty" db:"default_value"`
	Constraints  string `json:"constraints,omitempty"`
	Opts         string `json:"opts,omitempty"`
}
```
每一个`fieldinfo`都对应了可能需要用户输入的**参数**。
1. `Name`表示这个参数的名称，后续如果是以环境变量的方式传给s2ijob的话需要用这个Name。
2. `FieldType`表示这个参数的类型，它是一个以逗号分隔的属性，第一个表示参数的UI显示类型(仅供前端参考，具体是什么元素还需前端和设计考虑一下)，下表给出了我理解的对应关系：
   
    | 类型     | web元素                 |
    | -------- | ----------------------- |
    | text     | <input type='text'>     |
    | select   | <select>                |
    | checkbox | <input type="checkbox">checkbox在json中对应true和false |

    如果有第二个参数，表明了这个参数应该以何种方式传递给s2ijob，现在有下面几种形式：
    + environment： 以环境变量的形式，s2ijob有一个可以接受环境变量的参数。
    + tag: 放在builder_image后面，以tag的形式提供。这种用于让用户选择builder_image的版本，现在只有几个模板支持。
    如果有第三个参数（值为optional），表明了这个参数是可选参数，UI可以考虑放到可选参数一栏。

3. `DefaultValue`表明了这个参数的默认值
4. `Constraints`表示这个参数的限制，目前只有Source这个参数有一个`URL`的限制，表示用户输入的必须是一个合法的URL。
5. `Opts`用于FieldType是`select`的参数，表示其所有可能的options。

> 每一个模板都有一个`Source`参数，表示代码仓库的地址，创建s2ijob的时候是以`"source_url":"xxxxx"`传递过去的。

***

# 现在测试通过的语言框架

## Python
1. python3.5 + Django
   sample
```json
   {"source_url":"https://github.com/openshift/django-ex","builder_image":"centos/python-35-centos7","tag":"hello-python"}
```
输入参数只有两个，git地址和目标镜像
## Java
> 如果java是用maven编译的，可以添加`"reuse_maven_local_rep":true`参数保留maven本地依赖，下次继续运行这个任务的时候速度会明显提升。
1. openjdk 1.8+maven

sample
```json
{"source_url":"https://github.com/MagicSong/simple-java-maven-app.git","builder_image":"appuio/s2i-maven-java","tag":"hello-java"}
```
输入参数：
	+ 仓库地址
	+ 镜像名称
	+ （可选）mvn命令，默认`clean package`
	+ (可选) 是否跳过test 默认true

2. openjdk 1.8+maven+tomcat
sample
```json
{"source_url":"https://github.com/daticahealth/java-tomcat-maven-example.git","builder_image":"sarcouy/s2i-tomcat:8.5-jdk8-mvn3.3.9","tag":"hello-java-tomcat","environment_variables":"WAR_NAME=java-tomcat-maven-example.war"}
```  
输入参数：
	+ 仓库地址
	+ 镜像名称
	+ war的名称，包含后缀 


## Nodejs
1. nodejs8+express（此例子适合大部分nodejs程序，不仅仅express）
sample
```json
{"source_url":"https://github.com/MagicSong/node-js-sample.git","builder_image":"fortinj66/centos7-s2i-nodejs:8.9.4","tag":"hello-nodejs"}
```
输入参数：
	+ 仓库地址
	+ 镜像名称
  
## Php
1. php 5.6
```json
{"source_url":"https://github.com/MagicSong/opsworks-demo-php-simple-app.git","builder_image":"rvhoyt/s2i-php-56","tag":"hello-php"}
```	