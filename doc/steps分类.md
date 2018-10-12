# Jenkinsfile Steps分类

## Jenkinsfile用法简单指南
一个Jenkinsfile主要由agent、stages和post组成，agent表示jenkinsfile中step执行的环境，在我们程序中，agent是一个container，用户可以定义每个stage需要的agent，即定义一个镜像。agent具有覆盖的功能，而且也可以定义在stages中，在内的agent会覆盖掉外围的agent。为了降低复杂程度，我们的Jenkinsfile编辑器只需要为每一个stage定义agent，每个agent还可以定义环境变量。stages由一系列stage组成，每一个stage由很多step组成。

post表明了所有步骤完成之后执行的动作，是一个特殊的stage，他可以根据前面Stages的执行结果执行不同的命令。

> ⭐️符号表示此参数是必须的。需要注意的是，有一些命令如果只想输入一个参数的话（只有一个参数是required），是可以忽略参数名的，如`archiveArtifacts "xx/xx.tar"，而不需要写archiveArtifacts artifacts "xx/xx.tar"，echo、git也是同样的道理

## SCM
1. 拉取代码（经常用于拉取非git代码，例如svn等等）
### 命令名：checkout
### 参数 类型：
    1. branches []string  
    2. credentialsId string
    3. ⭐️ ️url string

1. 拉取Git代码
### 命令名：git
### 参数 类型：
    1. branch string
    2. credentialsId string
    3. ⭐️url string
    4. changelog bool (default true) 表示是否显示在Jenkins的changelog中
    5. poll bool (default true) 表示是否支持让jenkins轮询

1. 拉取svn代码
### 命令名：svn
### 参数 类型：
    1. ⭐️url string

## 常规流程
1. 打印命令
### 命令名：`echo`
### 参数 类型：
    1. ⭐️message string
2. 执行shell命令或脚本
### 命令名：`sh`
### 参数 类型：
    1. ⭐️script string
3. 发送邮件
### 命令名：`mail`
### 参数 类型：
    1. to string 收件人
    2. from string 发件人
    3. cc string 抄送
    4. ⭐️body 正文
    5. ⭐️subject string 主题
    6. bcc string 密送
4. 更改当前目录
### 命令名：`dir`
### 参数 类型：
    1. ⭐️path string
 
5. 在容器中执行命令
### 命令名：`container`
### 参数 类型：
    1. ⭐️name string
    2. shell string


## 编译相关
1. 保存制品
### 命令名：`archiveArtifacts`
### 参数 类型：
    1. ⭐️artifacts string
2. Source2Image
### 命令名：`s2i`
**目前还在测试中**
3. 清理工作空间
### 命令名：`cleanWs`
### 参数 类型：
    1. cleanWhenAborted bool
    2. notFailBuild bool

## 审核
1. Input
### 命令名：`input`
### 参数 类型：
    1. ⭐️message string
    2. id string
    3. submitter string(逗号分隔，必须没有任何空格)
    4. submitterParameter string
    5. parameter 非常复杂的类型，TODO
