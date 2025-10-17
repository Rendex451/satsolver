#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: $0 <directory>" >&2
    exit 1
fi

directory="$1"

if [ ! -d "$directory" ]; then
    echo "Error: Directory '$directory' does not exist" >&2
    exit 1
fi

# Создаем CSV файл с заголовком, если он не существует
csv_file="results.csv"
if [ ! -f "$csv_file" ]; then
    echo "filename,elapsed_time_ms" > "$csv_file"
fi

# Обрабатываем каждый .cnf файл в директории
for file in "$directory"/*.cnf; do
    if [ ! -f "$file" ]; then
        continue
    fi
    
    # Получаем только имя файла без пути
    filename=$(basename "$file")
    
    # Запускаем программу и парсим результат
    result=$(go run satsolver -f "$file" -c mc)
    
    # Парсим elapsed time
    elapsed_time=$(echo "$result" | grep -oP 'Elapsed time: \s*\K[0-9]+\.[0-9]ms')
    
    # Если время найдено, записываем в CSV
    if [ -n "$elapsed_time" ]; then
        echo "$filename,$elapsed_time" >> "$csv_file"
    fi
done