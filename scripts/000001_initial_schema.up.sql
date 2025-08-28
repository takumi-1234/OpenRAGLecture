-- 000001_initial_schema.up.sql

-- Base tables without foreign keys
CREATE TABLE `semesters` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `start_date` datetime(3) NOT NULL,
  `end_date` datetime(3) NOT NULL,
  `note` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_semesters_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `external_id` varchar(128) DEFAULT NULL,
  `email` varchar(255) NOT NULL,
  `password_hash` varchar(255) DEFAULT NULL,
  `display_name` varchar(128) DEFAULT NULL,
  `role` enum('student','instructor','ta','admin') NOT NULL DEFAULT 'student',
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_email` (`email`),
  KEY `idx_users_deleted_at` (`deleted_at`),
  KEY `idx_users_external_id` (`external_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Tables with foreign keys
CREATE TABLE `courses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `code` varchar(50) NOT NULL,
  `title` varchar(255) NOT NULL,
  `semester_id` bigint unsigned NOT NULL,
  `instructor_id` bigint unsigned DEFAULT NULL,
  `description` text,
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_courses_deleted_at` (`deleted_at`),
  KEY `idx_courses_code` (`code`),
  KEY `idx_courses_semester_id` (`semester_id`),
  CONSTRAINT `fk_courses_instructor` FOREIGN KEY (`instructor_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_courses_semester` FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `enrollments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL,
  `course_id` bigint unsigned NOT NULL,
  `role` enum('student','auditor','ta') NOT NULL DEFAULT 'student',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_enrollment_user_course` (`user_id`,`course_id`),
  KEY `idx_enrollments_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_enrollments_course` FOREIGN KEY (`course_id`) REFERENCES `courses` (`id`),
  CONSTRAINT `fk_enrollments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `documents` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `course_id` bigint unsigned NOT NULL,
  `semester_id` bigint unsigned NOT NULL,
  `title` varchar(255) NOT NULL,
  `source_uri` varchar(1024) DEFAULT NULL,
  `doc_type` enum('slides','pdf','notes','webpage','other') DEFAULT 'pdf',
  `version` bigint NOT NULL DEFAULT '1',
  `checksum` varchar(128) DEFAULT NULL,
  `metadata` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_documents_course_id` (`course_id`),
  KEY `idx_documents_semester_id` (`semester_id`),
  KEY `idx_documents_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_documents_course` FOREIGN KEY (`course_id`) REFERENCES `courses` (`id`),
  CONSTRAINT `fk_documents_semester` FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `pages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `document_id` bigint unsigned NOT NULL,
  `page_number` int DEFAULT NULL,
  `section_title` varchar(255) DEFAULT NULL,
  `language` varchar(16) DEFAULT NULL,
  `text` longtext,
  `token_count` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_pages_document_id` (`document_id`),
  KEY `idx_pages_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_pages_document` FOREIGN KEY (`document_id`) REFERENCES `documents` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `chunks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `page_id` bigint unsigned NOT NULL,
  `document_id` bigint unsigned NOT NULL,
  `course_id` bigint unsigned NOT NULL,
  `semester_id` bigint unsigned NOT NULL,
  `chunk_index` int NOT NULL,
  `start_offset` int DEFAULT NULL,
  `end_offset` int DEFAULT NULL,
  `text` longtext NOT NULL,
  `token_count` int DEFAULT NULL,
  `embedding_id` varchar(128) DEFAULT NULL,
  `embedding_model_version` varchar(64) DEFAULT NULL,
  `vector_hash` varchar(128) DEFAULT NULL,
  `score_meta` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_chunk_page_index` (`page_id`,`chunk_index`),
  KEY `idx_chunks_deleted_at` (`deleted_at`),
  KEY `idx_chunks_course_semester` (`course_id`,`semester_id`),
  KEY `idx_chunks_embedding_id` (`embedding_id`),
  CONSTRAINT `fk_chunks_course` FOREIGN KEY (`course_id`) REFERENCES `courses` (`id`),
  CONSTRAINT `fk_chunks_document` FOREIGN KEY (`document_id`) REFERENCES `documents` (`id`),
  CONSTRAINT `fk_chunks_page` FOREIGN KEY (`page_id`) REFERENCES `pages` (`id`),
  CONSTRAINT `fk_chunks_semester` FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `questions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `query_id` char(36) NOT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `course_id` bigint unsigned DEFAULT NULL,
  `semester_id` bigint unsigned DEFAULT NULL,
  `raw_query` text NOT NULL,
  `expanded_query` text,
  `ambiguity_score` float DEFAULT NULL,
  `use_llm_expansion` tinyint(1) NOT NULL DEFAULT '0',
  `tracing_meta` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_questions_deleted_at` (`deleted_at`),
  KEY `idx_questions_query_id` (`query_id`),
  CONSTRAINT `fk_questions_course` FOREIGN KEY (`course_id`) REFERENCES `courses` (`id`),
  CONSTRAINT `fk_questions_semester` FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`),
  CONSTRAINT `fk_questions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `answers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `question_id` bigint unsigned NOT NULL,
  `response_text` longtext NOT NULL,
  `response_model` varchar(128) DEFAULT NULL,
  `response_params` json DEFAULT NULL,
  `response_stream_ref` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_answers_deleted_at` (`deleted_at`),
  KEY `idx_answers_question_id` (`question_id`),
  CONSTRAINT `fk_answers_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `answer_sources` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `answer_id` bigint unsigned NOT NULL,
  `chunk_id` bigint unsigned NOT NULL,
  `score` float NOT NULL,
  `rank` int NOT NULL,
  `extracted_snippet` text,
  PRIMARY KEY (`id`),
  KEY `idx_answer_sources_deleted_at` (`deleted_at`),
  KEY `idx_answer_sources_answer_id` (`answer_id`),
  CONSTRAINT `fk_answer_sources_answer` FOREIGN KEY (`answer_id`) REFERENCES `answers` (`id`),
  CONSTRAINT `fk_answer_sources_chunk` FOREIGN KEY (`chunk_id`) REFERENCES `chunks` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `feedbacks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `answer_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `thumbs_up` tinyint(1) DEFAULT NULL,
  `comment` text,
  `label` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_feedbacks_deleted_at` (`deleted_at`),
  KEY `idx_feedbacks_answer_id` (`answer_id`),
  CONSTRAINT `fk_feedbacks_answer` FOREIGN KEY (`answer_id`) REFERENCES `answers` (`id`),
  CONSTRAINT `fk_feedbacks_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;