# roma



A super lightweight jumpserver service developed using the Go language.


## Deployment

roma needs a server with a public IP as the server for the jumpserver service.
This server needs external network access to be able to access the target server you need to access.

### Docker

```shell
$ docker pull bitrec/roma:latest
$ mkdir -p ~/.roma/.ssh
$ docker run \
  -p 2222:2222 \
  -v ~/.roma:/root\
  -v ~/.roma/.ssh:/root/.ssh\
  --name roma -d roma:latest
```

### Binary file

Download the version you need from the [Release](https://gitea.bitrec.ai/roma/releases) page, decompress it to get the `roma` binary executable, and run it.

```shell
$ ./roma
starting ssh server on port 2222...
```

## How to use

### First Time  

After the roma service is started, an sshd service will be started on port `2222`. You can also set the startup port through `-p`.

After the service is started, you only need to use the `ssh` command to access the service.

```shell
$ ssh 127.0.0.1 -p 2222
root@127.0.0.1's password:
New Username: root█
Password: ******█
Confirm your password: ******█
Please login again with your new acount.
Shared connection to 127.0.0.1 closed.
```

The default user password for the first access is `newuser`, and then the command line prompts to create a new user. Follow the prompts to create a new `admin` account for the jumpserver service.

```shell
$ ssh root@127.0.0.1 -p 2222
root@127.0.0.1's password:
Use the arrow keys to navigate: ↓ ↑ → ← 
? Please select the function you need: 
  ▸ List servers
    Edit users
    Edit servers
    Edit personal info
    Quit
```

You can use it after logging in with your password again.

### Upload or download file server via jumpserver

If you want to upload or download file from the server via jumpserver, you can use the `scp` command in the following format:

```shell
$ scp -P 2222 ~/Desktop/README.md  kubo@jumpserver:ops@server2:~/Desktop/README.md
README.md                                        100% 9279    73.9KB/s   00:00
```

```shell
scp -P 2222 kubo@jumpserver:ops@server2:~/Desktop/video.mp4 ~/Downloads
video.mp4                           100%   10MB  58.8MB/s   00:00
```

Note the use of `:` after `kubo@jumpserver` plus the `key` and `username` of the server you need to transfer, and finally write the destination or source path.
Folder transfer is currently not supported. Please compress the file and upload or download it.
