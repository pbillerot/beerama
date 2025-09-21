

# docker ps -a | grep beerama | awk '{print $1}' | xargs docker rm -f
docker build --no-cache -t beerama .
# docker compose up -d --build 
docker image prune -f
docker compose up -d
