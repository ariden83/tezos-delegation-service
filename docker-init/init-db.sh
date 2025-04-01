#!/bin/bash
set -e

# Installation de Sqitch pour PostgreSQL
apt-get update && apt-get install -y sqitch libdbd-pg-perl

# Exécution des migrations avec vérification
cd /app/sqitch_pg && sqitch deploy --verify