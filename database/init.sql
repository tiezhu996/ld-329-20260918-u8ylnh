CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  major VARCHAR(120) NOT NULL,
  credit_score INT NOT NULL DEFAULT 80,
  credit_level VARCHAR(40) NOT NULL
);

CREATE TABLE IF NOT EXISTS skills (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  title VARCHAR(120) NOT NULL,
  category VARCHAR(40) NOT NULL,
  level_score INT NOT NULL,
  campus VARCHAR(40) NOT NULL,
  description TEXT NOT NULL,
  portfolio VARCHAR(160) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS needs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  title VARCHAR(120) NOT NULL,
  category VARCHAR(40) NOT NULL,
  campus VARCHAR(40) NOT NULL,
  expect_time VARCHAR(80) NOT NULL,
  budget_type VARCHAR(40) NOT NULL,
  description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS appointments (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  initiator VARCHAR(80) NOT NULL,
  responder VARCHAR(80) NOT NULL,
  pair_name VARCHAR(170) NOT NULL,
  exchange_time VARCHAR(80) NOT NULL,
  place VARCHAR(120) NOT NULL,
  agenda VARCHAR(500) NOT NULL DEFAULT '',
  -- pending 待确认 / confirmed_partial 单方已确认 / confirmed 双方已确认
  -- / withdrawn 已撤回（终态）/ cancel_requested 取消待处理 / cancelled 已取消（终态）
  status VARCHAR(40) NOT NULL DEFAULT 'pending',
  initiator_confirmed TINYINT(1) NOT NULL DEFAULT 0,
  responder_confirmed TINYINT(1) NOT NULL DEFAULT 0,
  slots_held TINYINT(1) NOT NULL DEFAULT 1,
  cancel_reason VARCHAR(300) NULL,
  cancel_by VARCHAR(80) NULL,
  cancel_decision VARCHAR(20) NULL,
  version INT NOT NULL DEFAULT 1,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 时间档占用账本：撤回 / 取消生效后立即删除对应行释放时间档。
CREATE TABLE IF NOT EXISTS appointment_slots (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  appointment_id BIGINT NOT NULL,
  user_name VARCHAR(80) NOT NULL,
  slot VARCHAR(80) NOT NULL,
  UNIQUE KEY uk_user_slot (user_name, slot),
  CONSTRAINT fk_slot_appointment FOREIGN KEY (appointment_id) REFERENCES appointments(id)
);

CREATE TABLE IF NOT EXISTS reviews (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  from_user VARCHAR(80) NOT NULL,
  to_user VARCHAR(80) NOT NULL,
  rating INT NOT NULL,
  content TEXT NOT NULL
);

INSERT INTO users(name, major, credit_score, credit_level) VALUES
('林澈', '新闻传播 2023', 91, '黄金导师'),
('孟野', '音乐表演 2022', 88, '白银协作者'),
('周芮', '统计学 2021', 93, '黄金导师');

INSERT INTO skills(user_id, title, category, level_score, campus, description, portfolio) VALUES
(1, '毕业照人像摄影', '摄影', 92, '东校区', '提供构图、修图和毕业季跟拍，可交换吉他入门课。', '12组校园人像作品'),
(2, '民谣吉他陪练', '乐器', 81, '西校区', '节奏型、弹唱和舞台经验分享，想找人拍宣传照。', '校园音乐节演出视频'),
(3, 'Python 数据分析', '编程', 88, '中心校区', 'pandas、可视化、论文数据清洗辅导。', '3份课程项目证书');
