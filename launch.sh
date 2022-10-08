# sh launch.sh {3 4 5 6 7...}

if [ "$1" == '' ]; then
	echo "no argument --> 3 containers is the default"
	NUM=3
else
	NUM=$1
fi

if [ $NUM -lt 2 ]; then
	echo argument too small choose between 2 and 9
	exit 1
fi

if [ $NUM -gt 9 ]; then
	echo argument too big choose between 2 and 9
	exit 1
fi

docker network create calculator-network
echo "Building decentralized calculator image"
docker build -t $"decentralized-calculator-app" ./calculator-server/.

for i in $(seq 1 $NUM); do
	PORT=$((8000 + i))
	echo "Launching decentralized calculator n°$i at port $PORT"
	docker run -id --rm --name $"decentra-calcu-$i" --net calculator-network -p $PORT:8000 -d $"decentralized-calculator-app" ./calculator-server $i $NUM
done
