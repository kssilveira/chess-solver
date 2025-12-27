declare -a boards=(
	"   k,    ,P   ,KRNB"
	"   k,   p,P   ,KRNB"
	"b  k,   p,P   ,KRNB"
	"br k,   p,P   ,KRNB"
	"brnk,   p,P   ,KRNB"
)

for board in "${boards[@]}"
do
	echo -e "\nboard: '${board}'"
	time go run main.go --board="${board}" --enable_play=false --max_print_depth=-1 --print_depth=false --max_depth=2500000
done
