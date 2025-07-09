# ROMA
![Static Badge](https://img.shields.io/badge/License-AGPL_v3-blue)
![Static Badge](https://img.shields.io/badge/lightweight-green)

语言切换
[[🇨🇳 中文](README.md)]
[[🇺🇸 English](readme.res/README_en.md)]
[[🇷🇺 Русский](readme.res/README_ru.md)]
## 简介
#### ROMA 是一个AI驱动的使用 Go 语言开发的超轻量级跳板机服务，提供安全高效的远程访问解决方案。它支持多种资源类型，包括 Linux、Windows、数据库、路由器、交换机等，适用于各种运维场景。
---
![alt text](readme.res/face.png)

---
### TODO
- [ ] MCP支持（ai驱动）- 自动化运维
- [ ] Windows资源管理
- [ ] 数据库资源管理
- [ ] 路由器资源管理
- [ ] 交换机资源管理

## 功能特点

- **轻量级**：无需复杂配置，简单部署即可使用。
- **多资源支持**：支持 Linux、Windows、Docker、数据库、路由器、交换机等多种资源类型。
- **安全性**：使用 SSH 密钥认证，提高远程访问安全性。
- **简洁命令**：提供 `use`、`ls`、`ln` 等直观命令，简化操作。
- **用户管理**：支持 `whoami` 查询当前用户信息。
- **历史记录**：提供 `history` 命令，方便查看历史操作记录。
- **MCP支持**： model context protocol支持，支持自动化运维

---

## 安装与使用

### 1. 下载并编译

```sh
git clone https://github.com/bitrecAi/roma.git
cd roma
go build -o roma
```
### 2. 密钥配置
```toml
title = 'Roma Configs File'

[api]
gin_mode = 'release'
host = '0.0.0.0'
port = '6999'

[common]
language = 'zh'
port = '2200'
prompt = 'roma'

[database]
cdb_url = '/usr/local/roma/c.db'
rdb_passwd = ''
rdb_url = ''

[log]
level = 'debug'

[apikey]
prefix = 'apikey.'
key = 'AAAA2EAAHBZY26A25wOraC1c--------------------------xxx'    #接口用到的密钥

[user_1st]
email = 'super@test.x'
name = '超级管理员'
nickname = 'Super'
password = 'super001.'
public_key = '#<超级用户的私钥>'
username = 'super'
roles = "super,system,ops,ordinary,trial"

[control_passport]
service_user = 'root'
password = ''
resource_type = 'linux'
passport_pub = '<#跳板机的公钥>'
passport = '''<#跳板机的私钥>
'''
description = "default control's passport , and ops use this passport"

[banner]
show = true
banner = '''
       ______
      /\     \
     />.\_____\
   __\  /  ___/__        _ROMA__
  /\  \/__/\     \  ____/
 /O \____/*?\_____\
 \  /    \  /     /                 [A seamless solution for remote access, ensuring both efficiency and security.]
  \/_____/\/_____/
'''


#多角色设计
[[roles]]
name = "super"
desc = "all permissions [operation:user.(add|delete|update|get|list)]"

[[roles]]
name = "system"
desc = "system administrator [operation:resource.(add|delete|update|get|list)]"

[[roles]]
name = "ops"
desc = "system operations personnel [operation:resource.(get|list|use)]"

[[roles]]
name = "ordinary"
desc = "system ordinary [operation:resource-(*peripheral).(get|list)]"

[[roles]]
name = "trial"
desc = "system trial [operation:resource-(*trial).(get|list|use)]"
```

### 3. 运行

```sh
./roma
```

###

## 🔗 开源许可证
本项目基于 **GNU Affero General Public License (AGPL) v3.0** 开源发布。

📢 **重要**：
- 任何基于 ROMA 代码修改后用于提供**远程访问服务**的组织或个人，必须**开源他们的修改版本**。
- 详情请查看 [LICENSE](./LICENSE) 文件。