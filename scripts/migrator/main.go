// scripts/migrator/main.go
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/mysql"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
	"gorm.io/gorm"
)

// このプログラムは、GORMのAutoMigrate機能を使って
// データベーススキーマをGoのモデル定義に基づいて作成・更新します。
// `make test-e2e` コマンドから、プロジェクトのルートディレクトリを
// カレントワーキングディレクトリとして実行されることを想定しています。

func main() {
	// 1. プロジェクトルートにある .env ファイルを読み込む
	if err := godotenv.Load("./.env"); err != nil {
		log.Printf("Warning: could not load .env file from './.env'. Relying on environment variables. Error: %v", err)
	}

	// 2. configs ディレクトリから設定を読み込む
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load configuration from './configs': %v", err)
	}

	// 3. E2Eテスト用のDB接続情報を設定する
	// このスクリプトはコンテナ外のテストランナーから実行されるため、
	// Dockerホストの公開ポートに接続する必要がある。
	// ホストは常に 'localhost' で、ポートは .env の MYSQL_HOST_PORT を使用する。
	cfg.Database.MySQL.Host = "localhost"
	hostPort := os.Getenv("MYSQL_HOST_PORT")
	if hostPort == "" {
		hostPort = "3406" // .envにない場合のデフォルト値
		log.Printf("Warning: MYSQL_HOST_PORT not set in .env. Defaulting to %s", hostPort)
	}
	cfg.Database.MySQL.Port = hostPort

	// 4. DBに接続
	db, err := mysql.NewGORMClient(cfg.Database.MySQL)
	if err != nil {
		log.Fatalf("Failed to connect to database for migration: %v", err)
	}

	// 5. マイグレーションを実行
	runAutoMigration(db)
}

func runAutoMigration(db *gorm.DB) {
	log.Println("Running GORM Auto Migration...")
	err := db.AutoMigrate(
		&model.Semester{},
		&model.User{},
		&model.Course{},
		&model.Enrollment{},
		&model.Document{},
		&model.Page{},
		&model.Chunk{},
		&model.Question{},
		&model.Answer{},
		&model.AnswerSource{},
		&model.Feedback{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("GORM Auto Migration completed successfully.")
}
