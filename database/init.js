// MongoDB 初始化脚本：容器首次启动时自动执行，创建应用读写账号。
db = db.getSiblingDB('onlineexam_db');
db.createUser({
  user: 'onlineexam_user',
  pwd: 'onlineexam_pwd',
  roles: [{ role: 'readWrite', db: 'onlineexam_db' }]
});
