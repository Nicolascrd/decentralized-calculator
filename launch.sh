# sh launch.sh {3 4 5 6 7...} {0 1 2 ...}
# first arg : number of nodes (containers)
# second arg : number of failing nodes (returning random int instead of calculation)

if [ "$1" == '' ]; then
	echo "no first argument --> 3 containers is the default"
	NUM=3
else
	NUM=$1
fi

if [ "$2" == '' ]; then
	echo "no second argument --> 0 faulty containers is the default"
	BYZ=0
else
	BYZ=$2
fi

if [ $NUM -lt 2 ]; then
	echo first argument too small choose between 2 and 99
	exit 1
fi

if [ $NUM -gt 99 ]; then
	echo first argument too big choose between 2 and 99
	exit 1
fi

if [ $BYZ -lt 0 ]; then
	echo second argument too small choose between 0 and your first argument
	exit 1
fi

if [ $BYZ -gt $NUM ]; then
	echo second argument too big choose between 0 and your first argument
	exit 1
fi

docker network create calculator-network
echo "Building decentralized calculator image"
docker build -t $"decentralized-calculator-app" ./calculator-server/.

for i in $(seq 1 $NUM); do
	PORT=$((8000 + i))
	if [ $i -le $BYZ ]; then
		echo "Launching decentralized calculator n°$i at port $PORT, is byzantine"
		docker run -id --rm --name $"decentra-calcu-$i" --net calculator-network -p $PORT:8000 -d $"decentralized-calculator-app" ./calculator-server $i $NUM true
	else
		echo "Launching decentralized calculator n°$i at port $PORT, not byzantine"
		docker run -id --rm --name $"decentra-calcu-$i" --net calculator-network -p $PORT:8000 -d $"decentralized-calculator-app" ./calculator-server $i $NUM false
	fi
done
