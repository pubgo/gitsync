# gitsync

> 同步一个git仓库指定时间之前的代码到github中，或者到其他的git仓库中


## 思路

1. 启动的时候，检查，代码是否拉取，没有的话那么开始拉取代码，拉取之后的，并设置另一个remote origin 标记O1， 然后更新代码到最新
2. 获取两个月之前的改天的所有的需要提交的commit，并获取id，时间和msg
3. 获取距离两个月之前而当time最近的那一次commit的信息 标记为C1
4. git reset--hard C1.id
5. git reset--soft C1.id 的上一个CID
6. git commit -m "C1.msg"
6. git push O1 O1/branch

## 启动

## 创建远程仓库

1. 创建一个空的远程仓库，并放到配置文件当中
2. 没有创建远程仓库会提醒报错，并退出 

### 拉取依赖
`go mod vendor`

### 编译
`make b`

### 加密自己的密码
`./main ss --enc -k 秘钥 -t git仓库密码`

### 把加密后的密码配置到环境变量中
### 运行
`./main sync`

