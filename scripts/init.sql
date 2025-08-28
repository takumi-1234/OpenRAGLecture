-- scripts/init.sql

-- このSQLは、docker-composeでMySQLコンテナが初回起動する際に一度だけ実行されます。

-- 1. データベース（スキーマ）が存在しない場合に作成します。
CREATE DATABASE IF NOT EXISTS `open_rag_lecture` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 2. アプリケーションが使用するユーザーが存在しない場合に作成します。
--    パスワードは.envファイルで管理するのが望ましいですが、ここでは例として直接記述します。
CREATE USER IF NOT EXISTS 'user'@'%' IDENTIFIED BY 'password';

-- 3. 作成したユーザーに、対象データベースへの全ての権限を付与します。
GRANT ALL PRIVILEGES ON open_rag_lecture.* TO 'user'@'%';

-- 4. 権限設定を即座に反映させます。
FLUSH PRIVILEGES;