# DDNS-Updater - 传奇服务端动态IP自动更新工具

## 解决什么问题？

你的宽带是动态IP，每隔几天会换一次IP。每次换IP后你需要手动：
1. 打开引擎控制台 → 修改外网IP
2. 打开微端网关 → 修改微端地址
3. 重启相关服务

**这个工具帮你全自动完成！** IP一变，秒级自动更新，你什么都不用管。

## 工作原理

```
每60秒检测公网IP → 发现IP变化 → 自动更新引擎配置 → 自动更新微端数据库 → 自动执行重启命令
```

## 使用方法

### 第一步：下载

从 [Releases](../../releases) 页面下载 `ddns-updater.exe`（64位系统）或 `ddns-updater-x86.exe`（32位系统）

### 第二步：生成配置文件

把 `ddns-updater.exe` 放到一个固定目录（比如 `D:\DDNS-Updater\`），然后双击运行一次，或在命令行执行：

```cmd
ddns-updater.exe init
```

会自动生成 `config.yaml` 配置文件。

### 第三步：编辑配置文件

用记事本打开 `config.yaml`，修改以下内容：

#### 1. 文件路径（改成你的实际路径）

```yaml
file_updaters:
  - name: 引擎控制台-外网IP
    path: 'D:\MirServer\Mir200\Envir\MapQuest.txt'  # ← 改成你的路径
    old: '111.111.111.111'  # ← 第一次手动改成当前IP，之后自动更新
    new: '{{.IP}}'
```

#### 2. 数据库路径（微端网关）

```yaml
db_updaters:
  - name: 微端网关-服务器地址
    path: 'D:\MirServer\微端网关\wd.db'  # ← 改成你的路径
    queries:
      - sql: "UPDATE server_list SET address = '{{.IP}}'"
```

#### 3. 自动重启命令

```yaml
commands:
  - name: 重启微端网关
    cmd: 'cmd'
    args: ['/c', 'net', 'restart', 'WDGateway']  # ← 改成你的服务名
```

### 第四步：运行

```cmd
ddns-updater.exe
```

程序会在后台持续运行，自动检测IP变化并更新。

### 第五步：开机自启（推荐）

创建一个快捷方式放到启动目录：
```
Win+R → 输入 shell:startup → 回车 → 把快捷方式放进去
```

或者用任务计划程序：
1. 打开"任务计划程序"
2. 创建基本任务
3. 触发器：计算机启动时
4. 操作：启动程序 → 选择 `ddns-updater.exe`

## 命令说明

| 命令 | 说明 |
|------|------|
| `ddns-updater.exe` | 启动后台监控（默认） |
| `ddns-updater.exe init` | 生成默认配置文件 |
| `ddns-updater.exe check` | 立即检测一次IP并更新 |
| `ddns-updater.exe version` | 显示版本号 |

## 常见问题

### Q: 怎么知道IP变了？
程序每60秒检测一次公网IP（可配置），变化时会自动更新。控制台窗口会显示更新日志。

### Q: 支持哪些引擎？
翎风引擎、GEE引擎、V8引擎等主流传奇引擎都支持。只需要在配置文件中指定正确的文件路径和数据库路径。

### Q: 配置文件中的 `old` 字段怎么填？
第一次使用时，手动填入你当前的公网IP。之后程序会自动维护。如果留空占位符如 `YOUR_IP`，需要先手动改成实际IP。

### Q: 微端网关的数据库表名不对怎么办？
不同引擎的数据库结构可能不同。你可以用 [DB Browser for SQLite](https://sqlitebrowser.org/) 打开微端网关的 `.db` 文件，查看实际的表名和字段名，然后修改配置文件中的 SQL 语句。

### Q: 怎么后台运行不显示窗口？
可以用 `start /min ddns-updater.exe` 以最小化方式运行，或使用任务计划程序设置为"不管用户是否登录"。

### Q: 日志在哪里？
程序运行时会在控制台显示日志。如需保存日志，可以用：
```cmd
ddns-updater.exe >> log.txt 2>&1
```

## 技术说明

- 语言：Go（编译为单个 exe，无依赖）
- 配置：YAML 格式
- IP检测：HTTP API（多个备用源）
- 数据库：SQLite3（微端网关）
- 平台：Windows 7/10/11/Server

## 编译

本项目使用 GitHub Actions 自动编译，推送到 main 分支即触发。

本地编译需要 Go 1.21+ 和 GCC（CGO for SQLite）：
```cmd
go build -ldflags="-s -w" -o ddns-updater.exe .
```

## License

MIT
