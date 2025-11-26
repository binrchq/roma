curl -X POST -H "Content-Type: application/json" -H "apikey:apikey.YOUR_API_KEY_HERE" -d '{
  "type":"linux",
  "data": [
    {
      "hostname": "server1",
      "port": 22,
      "ipv4_pub": "203.0.113.10",
      "port_actual": 22,
      "ipv4_priv": "192.168.1.10",
      "ipv6": "",
      "port_ipv6": 22,
      "password": "",
      "username": "",
      "private_key": "",
      "description": "Example Linux configuration"
    },
    {
      "hostname": "server2",
      "port": 22,
      "ipv4_pub": "",
      "port_actual": 22,
      "ipv4_priv": "192.168.1.20",
      "ipv6": "",
      "port_ipv6": 22,
      "password": "",
      "username": "",
      "private_key": "",
      "description": "Another Linux configuration"
    }
  ]
}' http://127.0.0.1:6999/api/resource/add
