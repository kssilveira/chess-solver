set -o xtrace

time go run main.go --board='   k,    ,P   ,KR  '
time go run main.go --board='   k,    ,P   ,KR  ' --enable_promotion=true
time go run main.go --board='   k,    ,P   ,KR  ' --enable_drop=true
time go run main.go --board='   k,    ,P   ,KR  ' --enable_promotion=true --enable_drop=true

time go run main.go --board='   k,    ,P   ,KRNB'
time go run main.go --board='   k,    ,P   ,KRNB' --enable_promotion=true

time go run main.go --board='   k,   p,P   ,KRNB'
time go run main.go --board='   k,   p,P   ,KRNB' --enable_promotion=true

time go run main.go --board='b  k,   p,P   ,KRNB'
