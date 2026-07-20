-- Migration: 001_init.down.sql
-- Drop initial schema for Vert Helper

DROP TABLE IF EXISTS worker_snapshots CASCADE;
DROP TABLE IF EXISTS workers CASCADE;
DROP TABLE IF EXISTS action_executions CASCADE;
DROP TABLE IF EXISTS questions CASCADE;
DROP TABLE IF EXISTS actions CASCADE;
DROP TABLE IF EXISTS service_health CASCADE;
DROP TABLE IF EXISTS services CASCADE;
