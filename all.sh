declare -a boards=(
	"   k,    ,P   ,KRNB"
	"   k,   p,P   ,KRNB"
	"b  k,   p,P   ,KRNB"
	"br k,   p,P   ,KRNB"
	"brnk,   p,P   ,KRNB"
)

declare -a promotions=(
	"false"
	"true"
)

for board in "${boards[@]}"
do
	for promotion in "${promotions[@]}"
	do
		echo -e "\n--board='${board}' --enable_promotion=${promotion}"
		time go run main.go --board="${board}" --enable_play=false --max_print_depth=-1 --print_depth=true --enable_promotion="${promotion}"
	done
done
