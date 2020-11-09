## web hook 自动构建


## 使用
* `git clone https://github.com/seth-shi/web-hook`
* 按需修改`.env`配置
* `cp .env.example .env`
* `按项目配置 hooks.json`
* `cp _hooks.json hooks.json`
* 构建然后执行可执行文件
* `go build`
****

## `hooks.json 说明` (钉钉需要在自定义关键词增加`build`关键字,方可发送消息成功)
```shell
{
    # 项目名字,用于路由 /hooks/:name 匹配到项目
    # 如访问 http://xx.com/test 则会触发以下构建
    "name": "test",
    # 构建项目根目录
    "dir": "./",
    # 通知
    "notifications": [
      {
         # dingtalk 写死
        "type": "dingtalk",
        # 钉钉的 webhook 地址, 需在机器人设置
        # 自定义关键词: build
        "web_hook": "https://oapi.dingtalk.com/robot/send?access_token="
      }
    ],
    # 某些 web hook 不想触发(如 master)
    # 一下是过滤条件
    "hook_filters": [
      {
        # 参数形式的 key value
        "type": "parameters",
        "key": "ref",
        "value": "refs/heads/dev"
      },
      # header 形式的 key value
      {
        "type": "header",
        "key": "X-Gitlab-Event",
        "value": "Push Hook"
      }
    ],
    # pull 代码之后触发的构建脚本
    "hooks": [
      {
        # 执行的环境目录
        "dir": "/var/www",
        # 可执行的 shell
        "shell": "docker -v",
        # 断言执行 shell 的输出包含内容才才成功
        "assert": "Docker version",
        # 断言执行 shell 的不输出包含内容才才成功
        "assert_no": "",
        # 断言失败, 是否执行下一个
        "assert_fail_continue": false
      },
      {
              "dir": "/var/www",
              "shell": "composer install",
              "assert": "",
              "assert_no": "Composer could not find a composer.json",
              "assert_fail_continue": false
      }
    ],
    # 发生错误的时候执行的脚本,用于清理工作
    "fail_hooks": [
      {
        "dir": "/var/www",
        "shell": "composer dump-auto"
      }
    ]
  }
```