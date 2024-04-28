API=http://127.0.0.1:5666
MIP_TEST_HOST=yd-guangxi-liuzhou-sn6-117-141-19-202
PPPOE_TEST_HOST=yd-jiangsu-suzhouchangshu-sn2-172-31-3-13

run:
	go run cmd/roma/main.go

run_roma:
	go run cmd/roma/main.go -f

test:
	go run cmd/test/main.go
	
request_root:
	@curl ${API}/

# 备份数据库文件
clear_db:
	mv /usr/local/roma/c.db /usr/local/roma/c.db.bk

# 清理指定端口上的进程
clear_port_processes:
	$(eval PORT := $(filter-out $@,$(MAKECMDGOALS)))
	@lsof -ti:$(PORT) | xargs kill -9

# 清理 2222 和 6999 端口上的进程
clear:
	make clear_port_processes 2222
	make clear_port_processes 6999

# 清理所有
clear_all: clear clear_db

