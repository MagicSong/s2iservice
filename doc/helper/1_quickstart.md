# 一、快速开始

## 1. 创建DevOPS项目
想要使用Kubesphere的DevOps功能，首先就要创建一个DevOps工程。DevOps工程位于企业空间中，一个企业空间中可以有多个DevOps工程。本节会介绍如何创建一个DevOPS项目，并进行相关的成员配置。

## 2. 创建流水线
### Jenkinsfile in SCM
> Jenkinsfile in SCM指的是将Jenkinsfile本身也放入SCM（Source Control Management，源代码管理），其他使用`git clone`或者其他类似的命令都能够获取到这个Jenkinsfile，从而能快速使用项目的CI/CD功能。这种模式比较常用。

1. 进入DevOPS项目，进入左边菜单中的流水线。点击右边的创建按钮。
   
   ![create](images/pipeline_create.png)

2. 在跳出的弹框中，输入Pipelined的基本信息，输入完成，不要点击下一步。
   
   ![input_info](images/pipeline_info.png)

3. 点击代码仓库，选择Github，并输入访问仓库用的access token。
   
   ![git_input](images/pipeline_git_token.png)

4. 如果输入token没有错误，在右边会跳出用户的所有代码库，在其中选择一个带有Jenkinsfile的仓库。如果没有此类的仓库，可以是使用我们的测试仓库<https://github.com/MagicSong/simple-java-maven-app.git>。将其Fork到您的仓库中，然后就会在右边的列表中显示。
   
   ![git_select](images/git_repo.png)

5. 完成代码的设置之后，进入高级设置。这里直接跳过，高级设置可参考[3. 高级配置](#3-高级配置)，点击下一步完成Pipeline的创建。
   
   ![advance](images/pipeline_advance.png)

6. 点击扫描分支，用于获取源代码中的Jenkinsfile。扫描分支会获取仓库的所有分支，并且根据输入的路径寻找Jenkinsfile（默认是根目录下一个名为Jenkinsfile的文件）。等待一段时间，Pipeline就会运行完成。至此就完成了一个Jenkinsfile in SCM的创建。
   
   ![branch_scan](images/pipeline_scan.png)
   
7. 点击页面中运行序列号，查看运行的具体状态。页面展现了Pipeline中每一个步骤的运行状态，黑色框标注了Pipeline的步骤名称。 
   
   ![run_status](images/pipeline_status.png)

8. 当前页面中点击右上方的`查看日志`，查看Pipeline运行的日志。页面展示了每一步的具体日志、运行状态以及时间等信息。日志可以下载到本地，如果出现了错误，下载到本地更便于分析。
   
   ![log](images/pipeline_log.png)

##v# Jenkinsfile out of SCM
> `Jenkinsfile out of SCM`不同于上面，它的代码仓库中不存储Jenkinsfile，用户需要另外提供Jenkinsfile。由于这种模式代码仓库不存储CI/CD信息，后期维护不方便，所以相比上面一种用的少。

1. 进入一个DevOPS项目，创建一个Pipeline。创建步骤和上面的一样，但是要注意这里不需要选择仓库，让仓库选项为空。
   ![select_git](images/pipeline_outscm_git.png)

2. 点击下一步，同样跳过高级设置（这里的高级设置和上面的有一些差别，因为类型不一样）。点击创建完成Pipeline的创建

   ![create](images/pipeline_outscm_create.png)

3. 创建完成之后页面就会自动跳转到编辑Jenkinsfile的页面，在此页面中通过创建Pipeline的每个步骤执行的内容就可以生成Jenkinsfile，用户无需学习Jenkinsfile的语法，非常方便。

   ![pipeline_editor](images/pipeline_editor.png)

4. 点击左侧结构编辑区域的加号，增加一个Stage。然后在Stage里点击`增加步骤`,如下图。
   
   ![add_stage](images/pipeline_outscm_editor_addstage.png)

5. 在右边的内容编辑区域，选择git，在URL中输入代码仓库地址，如果没有现成可用的仓库，可以输入<https://github.com/MagicSong/devops.git>做测试用。点击确定之后，给这个Stage取一个名称，可以任意指定，也可以按照下图输入。

   ![add_scm](images/pipeline_outscm_editor_addscm.png)

6. 然后继续在左边点击+号增加一个Stage，这个Stage用于编译代码。这个操作需要在容器中进行，添加一个`container`的步骤，按照下图进行输入，最后还需给这个Stage取名。
   
   ![add_build](images/pipeline_outscm_editor_addbuild.png)

7. 编译完成之后，就可以打包镜像了，打包只需要在shell中执行命令就可以了，按照下图添加一个Stage。

   ![add_docker](images/pipeline_outscm_editor_adddocker.png)

8. 点击保存，这样就完成了一个Jenkinsfile的创建。