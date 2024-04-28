#! /bin/bash

#如果dasel命令不存在，则安装dasel
if ! command -v dasel &>/dev/null; then
    #安装dasel
    curl -sSLf "$(curl -sSLf https://api.github.com/repos/tomwright/dasel/releases/latest | grep browser_download_url | grep linux_amd64 | grep -v .gz | cut -d\" -f 4)" -L -o dasel && chmod +x dasel
    mv ./dasel /usr/local/bin/dasel
fi
#判断生产环境的配置文件是否存在
CONFIG_FILE="/etc/roma/config.toml"
if [ ! -f $CONFIG_FILE ]; then
    CONFIG_FILE="configs/config.toml"
    if [ ! -f $CONFIG_FILE ]; then
        echo "配置文件不存在"
        exit 1
    fi
fi

#读取配置文件的主KEY
ROMA_ASCII01='''       ______
      /\     \
     />.\_____\
   __\  /  ___/__        _ROMA__
  /\  \/__/\     \  ____/
 /O \____/*?\_____\
 \  /    \  /     /                 [A seamless solution for remote access, ensuring both efficiency and security.]
  \/_____/\/_____/
'''
echo "$ROMA_ASCII01"
echo "ROMA Config File: $CONFIG_FILE"
echo "begin--------------------------------------------------->"
KEYS=($(dasel -f $CONFIG_FILE --pretty -r toml "keys()" | sed "s/[][]//g; s/'//g; s/,/ /g"))
for KEY in "${KEYS[@]}"; do
    if [ "$KEY" == "title" ] || [ "$KEY" == "version" ] || [ "$KEY" == "roles" ]; then
        continue
    fi
    echo -e "\e[34m[Module\e[0m \e[33m$KEY\e[0m \e[34mConfig]\e[0m"
    KEYS_2ND=($(dasel -f $CONFIG_FILE --pretty -r toml "$KEY.keys()" | sed "s/[][]//g; s/'//g; s/,/ /g"))
    for KEY_2ND in "${KEYS_2ND[@]}"; do
        default_value=$(dasel -f $CONFIG_FILE --pretty -r toml "$KEY.$KEY_2ND")
        echo -ne "\e[32m请输入$KEY_2ND(default:$default_value):\e[0m"
        read -r input
        if [ -z "$input" ]; then
            input="$default_value"
        else
            dasel -f $CONFIG_FILE put -r toml -t string -v "$input" $KEY.$KEY_2ND
            if [ $? -ne 0 ]; then
                echo -e "\e[31m$KEY_2ND:$input fail\e[0m"
                exit 1
            fi
        fi
    done
    echo ""
done
