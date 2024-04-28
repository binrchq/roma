curl -X POST -H "Content-Type: application/json" -H "apikey:apikey.AAAA2EAAHBZY26A25wOraC1c2E2BxkKFrNmWoLMOXwWVnwFv5s7Q8w" -d '{
  "type":"linux",
  "data": [
    {
      "hostname": "hk2",
      "port": 22,
      "ipv4_pub": "159.138.2.226",
      "port_actual": 22,
      "ipv4_priv": "10.2.0.6",
      "ipv6": "",
      "port_ipv6": 22,
      "password": "",
      "username": "",
      "private_key": "",
      "description": "Example Linux configuration"
    },
    {
      "hostname": "js",
      "port": 22,
      "ipv4_pub": "",
      "port_actual": 22,
      "ipv4_priv": "10.2.0.4",
      "ipv6": "",
      "port_ipv6": 22,
      "password": "",
      "username": "",
      "private_key": "",
      "description": "Another Linux configuration"
    }
  ]
}' http://127.0.0.1:6999/api/resource/add
