## 免责声明

本教程仅为合法的教学目的而准备，严禁用于任何形式的违法犯罪活动及其他商业行为，在使用本教程前，您应确保该行为符合当地的法律法规，继续阅读即表示您需自行承担所有操作的后果，如有异议，请立即停止本文章阅读。

## 前言

本人用的是雷池WAF开心版，但是很多功能被压制了，很想支持一下雷池WAF，但是奈何年费3600太贵了，但凡少一个0，我就支持了。于是我就开始了自动化实现运营的操作。

## 映射数据库端口

```
#!/bin/bash

# 运行安装/更新脚本
bash -c "$(curl -fsSLk https://waf-ce.chaitin.cn/release/latest/upgrade.sh)"

# 进入 /data/safeline 目录
cd /data/safeline || { echo "/data/safeline not found!"; exit 1; }

# 检查 compose.yaml 是否存在并备份
if [ -f compose.yaml ]; then
    echo "Backing up the current compose.yaml"
    cp compose.yaml compose.yaml.bak
else
    echo "compose.yaml not found in /data/safeline!"
    exit 1
fi

# 检查是否已经存在端口映射
if grep -q "5433:5432" compose.yaml; then
    echo "PostgreSQL port mapping already exists."
else
    # 使用 sed 插入 ports 字段到 postgres 服务中
    sed -i '/container_name: safeline-pg/a\    ports:\n      - 5433:5432' compose.yaml
    echo "PostgreSQL port mapping added to 5433:5432."
fi

# 重新启动容器，应用更改
docker compose down --remove-orphans && docker compose up -d

echo "Containers restarted with the updated compose.yaml"
```

这个脚本适用于每次更新时，重新映射数据库端口。

## 提取告警日志

通过`cat /data/safeline/.env | grep POSTGRES_PASSWORD | tail -n 1 | awk -F '=' '{print $2}'`查看数据库密码

然后在`/var/scripts/.pgpass`中写入如下代码，然后给这个文件添加600的权限。

```
localhost:5433:safeline-ce:safeline-ce:abcd #把abcd替换成第2步中获取到的密码
```

```
#!/bin/bash

# 设置PGPASSFILE环境变量
export PGPASSFILE=/var/scripts/.pgpass

# PostgreSQL 的连接信息
PG_HOST="localhost"
PORT="5433"
DATABASE="safeline-ce"
USERNAME="safeline-ce"
HOSTNAME=$(hostname)
PROGRAM_NAME="safeline-ce"

#获取最后一条WAF攻击事件日志的ID，日志数据存储在MGT_DETECT_LOG_BASIC表中
LAST_ID=$(psql -h $PG_HOST -p $PORT -U $USERNAME -d $DATABASE -t -P footer=off -c "SELECT id FROM PUBLIC.MGT_DETECT_LOG_BASIC ORDER BY id desc limit 1")
while true;do
#从pgsql数据库中获取waf的最新攻击事件日志，如果没有产生新日志，这条SQL会返回空
    raw_log=$(psql -h $PG_HOST -p $PORT -U $USERNAME -d $DATABASE -t -P footer=off -c "SELECT TO_CHAR(to_timestamp(timestamp) AT TIME ZONE 'Asia/Shanghai', 'YYYY-MM-DD HH24:MI:SS'), CONCAT(PROVINCE, CITY) AS SRC_CITY, SRC_IP, CONCAT(HOST, ':', DST_PORT) AS HOST,url_path,rule_id,id FROM PUBLIC.MGT_DETECT_LOG_BASIC where id > '$LAST_ID' ORDER BY id asc limit 1")
#检查SQL查询结果，如果有新加的日志就执行以下操作，把SQL查询结果重写为syslog日志，并记录到/var/log/waf_alert/waf_alert.log
    if [ -n "$raw_log" ]; then
        ALERT_TIME=$(echo "$raw_log" | awk -F ' \\| ' '{print $1}')
        SRC_CITY=$(echo "$raw_log" | awk -F ' \\| ' '{print $2}')
        SRC_IP=$(echo "$raw_log" | awk -F ' \\| ' '{print $3}')
        DST_HOST=$(echo "$raw_log" | awk -F ' \\| ' '{print $4}')
        URL=$(echo "$raw_log" | awk -F ' \\| ' '{print $5}')
        RULE_ID=$(echo "$raw_log" | awk -F ' \\| ' '{print $6}')
        EVENT_ID=$(echo "$raw_log" | awk -F ' \\| ' '{print $7}')
        syslog="src_city:$SRC_CITY, src_ip:$SRC_IP, dst_host:$DST_HOST, url:$URL, rule_id:$RULE_ID, log_id:$EVENT_ID"
        echo $ALERT_TIME $HOSTNAME $PROGRAM_NAME: $syslog >> /var/log/waf_alert/waf_alert.log
#更新最后处理的事件ID
        LAST_ID=$(($LAST_ID+1))
    fi
    sleep 3
done
```

将告警日志提取到`/var/log/waf_alert/waf_alert.log`中，执行日志提取脚本前请先安装`apt install postgresql-client`

放入后台执行，或写入systemctl开机启动项`nohup /var/scripts/waf_log.sh > /dev/null 2>&1 &`

## 监控日志文件变化

监控日志文件变化，发生变化，则触发告警脚本，告警脚本默认提取最后一条告警来发送钉钉通知

为了实现持久化监控日志文件编号并执行特定命令，可以使用Linux上的`inotify-tools`，特别是 `inotifywait` 命令。`inotifywait` 能够监控文件的修改事件，并在检测到变化时触发某条命令。

### 安装inotify-tools

```
apt update
apt install inotify-tools
```

### 使用inotify-tools

创建一个monitor.sh脚本，内容如下：

```
#!/bin/bash

# 监控的日志文件路径
LOG_FILE="/var/log/waf_alert/waf_alert.log"

# 定义检测到日志文件变化时执行的命令（可以根据需要修改）
COMMAND_TO_EXECUTE="/var/scripts/Log_Push_linux_amd64"	# 这里后续执行钉钉推送消息

# 使用inotifywait持续监控日志文件的修改
inotifywait -m -e modify "$LOG_FILE" | while read path action file; do
    echo "Detected $action on $file. Executing command..."
    # 执行指定的命令
    $COMMAND_TO_EXECUTE
done
```

这个脚本放入后台执行，或者设置一个开机自启项。

## 实现钉钉推送

这里我使用Golang来进行实现，不仅实现钉钉推送也可以进行Server酱推送。

下面是源码看附件ZIP

## 实现效果

### 钉钉
![0842d4229c75166647180b3b8d0825c](https://github.com/user-attachments/assets/473ab02c-ba70-4282-8b52-54d5d6db13ad)

### Server酱

Server酱的API调用暂时未完善!
