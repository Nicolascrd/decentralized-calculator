for i in 1 2 3
do
	PORT=$((8000+i)) 
	echo "Building decentralized calculator image n°$i at port $PORT"
	docker build -t $"decentralized-calculator-app-$i" --build-arg PORT=$PORT --build-arg NUM=$i ./calculator-server/.
	echo "Launching decentralized calculator n°$i at port $PORT"
	docker run -id --rm --name $"decentra-calcu-$i" -d -p $PORT:$PORT $"decentralized-calculator-app-$i" ./calculator-server $i
done

