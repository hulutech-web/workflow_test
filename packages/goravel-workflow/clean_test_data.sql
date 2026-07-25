-- 清理工作流测试数据
DELETE FROM entrydatas;
DELETE FROM proc_comments;
DELETE FROM proc_add_signs;
DELETE FROM cc_records;
DELETE FROM procs;
DELETE FROM entries;

-- 重置自增ID
ALTER TABLE entries AUTO_INCREMENT = 1;
ALTER TABLE procs AUTO_INCREMENT = 1;
ALTER TABLE entrydatas AUTO_INCREMENT = 1;
ALTER TABLE proc_comments AUTO_INCREMENT = 1;
ALTER TABLE proc_add_signs AUTO_INCREMENT = 1;
ALTER TABLE cc_records AUTO_INCREMENT = 1;
