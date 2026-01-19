#!/bin/bash
# Минимальный скрипт для бесконечной отправки рандомных команд
# Использует изображение 100x100 пикселей

IMAGE_URL="https://httpbin.org/image/png"
API_URL="http://localhost:8080/send"

random_num() {
    echo $((RANDOM % $2 + $1))
}

while true; do
    cmd_id="cmd-$(date +%s)-$RANDOM"
    
    case $((RANDOM % 6)) in
        0) # resize
            width=$(random_num 50 100)
            height=$(random_num 50 100)
            curl -s -X POST "$API_URL" \
                -H "Content-Type: application/json" \
                -d "{\"id\":\"$cmd_id\",\"command\":\"resize\",\"image_url\":\"$IMAGE_URL\",\"parameters\":{\"width\":$width,\"height\":$height}}" > /dev/null
            ;;
        1) # crop (гарантированно внутри 100x100)
            x=$(random_num 0 60)
            y=$(random_num 0 60)
            width=$(random_num 10 $((100 - x)))
            height=$(random_num 10 $((100 - y)))
            curl -s -X POST "$API_URL" \
                -H "Content-Type: application/json" \
                -d "{\"id\":\"$cmd_id\",\"command\":\"crop\",\"image_url\":\"$IMAGE_URL\",\"parameters\":{\"x\":$x,\"y\":$y,\"width\":$width,\"height\":$height}}" > /dev/null
            ;;
        2) # filter
            filters=("blur" "sharpen" "grayscale" "invert")
            filter="${filters[$((RANDOM % 4))]}"
            intensity=$(awk -v min=0.1 -v max=1.0 'BEGIN{srand(); printf "%.2f", min+rand()*(max-min)}')
            curl -s -X POST "$API_URL" \
                -H "Content-Type: application/json" \
                -d "{\"id\":\"$cmd_id\",\"command\":\"filter\",\"image_url\":\"$IMAGE_URL\",\"parameters\":{\"filter_type\":\"$filter\",\"intensity\":$intensity}}" > /dev/null
            ;;
        3) # transform
            rotation=$(random_num 0 360)
            flip_h=$((RANDOM % 2))
            flip_v=$((RANDOM % 2))
            curl -s -X POST "$API_URL" \
                -H "Content-Type: application/json" \
                -d "{\"id\":\"$cmd_id\",\"command\":\"transform\",\"image_url\":\"$IMAGE_URL\",\"parameters\":{\"rotation_degrees\":$rotation,\"flip_horizontal\":$flip_h,\"flip_vertical\":$flip_v}}" > /dev/null
            ;;
        4) # analyze
            models=("object-detection" "color-analysis")
            model="${models[$((RANDOM % 2))]}"
            curl -s -X POST "$API_URL" \
                -H "Content-Type: application/json" \
                -d "{\"id\":\"$cmd_id\",\"command\":\"analyze\",\"image_url\":\"$IMAGE_URL\",\"parameters\":{\"models\":[\"$model\"]}}" > /dev/null
            ;;
        5) # remove_background
            curl -s -X POST "$API_URL" \
                -H "Content-Type: application/json" \
                -d "{\"id\":\"$cmd_id\",\"command\":\"remove_background\",\"image_url\":\"$IMAGE_URL\",\"parameters\":{\"output_format\":\"png\",\"high_quality\":$((RANDOM % 2))}}" > /dev/null
            ;;
    esac
    
    echo -n "."
    sleep 0.5
done
