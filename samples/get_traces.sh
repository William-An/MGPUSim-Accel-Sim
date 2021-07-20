#!/usr/bin/env bash
# Generate and collect AMD GCN3 traces

already_done() { 
  local elapsed=${1}
  local done=0
  while [ "$done" -lt "$elapsed" ]; do
    printf "▇"; 
    done=$(($done + 1))
  done 
}
remaining() { 
  local elapsed=${1}
  local duration=${2}
  while [ "$elapsed" -lt "$duration" ]; do
    printf " "; 
    elapsed=$(($elapsed + 1))
  done 
}
percentage() { 
  local elapsed=${1}
  local duration=${2}
  printf "| %s%%" $(( (($elapsed)*100)/($duration)*100/100 )); 
}
clean_line() { 
  printf "\r";
  printf "%-100s" " "
  printf "\r"; 
}

cwd=$(readlink -f .)
cwd=${cwd##*/}

if [[ ! "$cwd" == "samples" ]]; then
  echo "This shell script must be run in MGPUSim/samples folder, exiting now"
  exit
fi

# Create traces dir
tracesDir=$(readlink -f ./traces/)  # absolute path to trace folder
if [[ ! -e $tracesDir ]]; then
  mkdir $tracesDir;
else
  rm ${tracesDir%/}/*.trace 2> /dev/null
  rm ${tracesDir%/}/kernelslist 2> /dev/null
  rm ${tracesDir%/}/*/*.trace 2> /dev/null
  rm ${tracesDir%/}/*/kernelslist 2> /dev/null
  rmdir $tracesDir/* 2> /dev/null
fi

# Loop benchmarks dir to generate traces
benchmarkDirs=*/
numBenchmarks=$((`echo $benchmarkDirs | wc -w` - 3))
curr=0
for benchmarkDir in $benchmarkDirs; do
  # Ignore runner dir
  if [[ "$benchmarkDir" == "runner/" || "$benchmarkDir" == "traces/" || "$benchmarkDir" == "server/" ]]; then
    continue
  fi
  benchmark=${benchmarkDir%/}
  # Progress bar
  curr=$((curr + 1))
  already_done $curr
  remaining $curr $numBenchmarks
  percentage $curr $numBenchmarks
  echo -ne " Getting traces for $benchmark"

  # Build benchmark and run to generate traces
  cd $benchmarkDir
  go build

  if [[ "$benchmark" == "xor" ]]; then  # XOR in parallel might caused deadlock issue
    ./$benchmark -debug-isa >> ../get_traces.log 2>> ../get_traces.err
  else
    ./$benchmark -parallel -debug-isa >> ../get_traces.log 2>> ../get_traces.err
  fi

  # Store generated traces
  benchmarkTraceDir=$tracesDir/$benchmark
  mkdir $benchmarkTraceDir
  mv *.trace $benchmarkTraceDir
  mv kernelslist $benchmarkTraceDir
  cd ..
  clean_line
done
echo "done"

