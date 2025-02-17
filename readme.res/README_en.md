# ROMA
![Static Badge](https://img.shields.io/badge/License-AGPL_v3-blue)
![Static Badge](https://img.shields.io/badge/lightweight-green)

language switching
[[🇨🇳 中文](README.md)]
[[🇺🇸 English](readme.res/README_en.md)]
[[🇷🇺 Русский](readme.res/README_ru.md)]

## Introduction
#### ROMA is an ultra-lightweight springboard service developed using Go language that provides a safe and efficient remote access solution. It supports multiple resource types, including Linux, Windows, databases, routers, switches, etc., and is suitable for various operation and maintenance scenarios.
---
![alt text](face.png)

---
### TODO
- [ ] Windows Explorer
- [ ] Database resource management
- [ ] Router resource management
- [ ] Switch resource management

##Functional characteristics
- **Lightweight**: No complex configuration is required and can be used easily by simple deployment.
- **Multi-resource support**: Supports multiple resource types such as Linux, Windows, Docker, database, router, switch, etc.
- **Security**: Use SSH key authentication to improve remote access security.
- **Concise commands**: Provide intuitive commands such as `use`,`ls`, and `ln` to simplify operations.
- **User management**: Support whoami to query current user information.
- **History**: Provide the `history` command to facilitate viewing of historical operation records.

---

## installation and use

### 1. Download and compile

```sh
git clone https://github.com/bitrecAi/roma.git
cd roma
go build -o roma
```
### 2. key configuration
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
key = 'AAAA2EAAHBZY26A25wOraC1c--------------------------xxx'    #Key used for interface

[user_1st]
email = 'super@test.x'
name = 'Super admin'
nickname = 'Super'
password = 'super001.'
public_key = '#<Superuser.s private key>'
username = 'super'
roles = "super,system,ops,ordinary,trial"

[control_passport]
service_user = 'root'
password = ''
resource_type = 'linux'
passport_pub = '<#The public key of the springboard machine>'
passport = '''<#Private key of the springboard machine>
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


#More role 
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

### 3. run

```sh
./roma
```

## Open source license
This project is released under the **GNU Affero General Public License (AGPL) v3.0** open source.

📢 **Important**:
- Any organization or individual whose ROMA-based code has been modified to provide remote access services must ** open source their modified version **.
- For details, please visit [LICENSE](./LICENSE) document.