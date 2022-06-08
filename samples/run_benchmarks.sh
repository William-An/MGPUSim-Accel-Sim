#!/usr/bin/env bash
# Run MGPUSim benchmarks

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

# Clean up log files
rm ./run_benchmarks.log
rm ./run_benchmarks.err

# Loop benchmarks dir to generate traces
benchmarkDirs=*/
numBenchmarks=$((`echo $benchmarkDirs | wc -w` - 3))
curr=0
runBenchmarks=""
pids=""
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
  echo " Launching $benchmark "

  # Build benchmark and run to generate traces
  cd $benchmarkDir
  go build

  if [[ "$benchmark" == "xor" ]]; then  # XOR in parallel might caused deadlock issue
    ./$benchmark -report-all -timing >> ../run_benchmarks.log 2>> ../run_benchmarks.err &
  else
    ./$benchmark -parallel -report-all -timing >> ../run_benchmarks.log 2>> ../run_benchmarks.err &
  fi
  pids="$pids $!"
  runBenchmarks="$runBenchmarks $benchmark"
  cd ..
  clean_line
done

# Check for job status every 5 sec
allDone=0
while [[ allDone -eq 0 ]]; do
    allDone=1
    notFinished=""
    for benchmark in $runBenchmarks; do
        # Check if benchmark is still running
        pgrep $benchmark 2>&1 > /dev/null

        if [[ $? -eq 0 ]]; then
            allDone=$((allDone & 0))
            notFinished="$notFinished $benchmark"
        else
            echo "$benchmark done"
        fi
    done
    runBenchmarks=$notFinished
    sleep 5
done

wait
echo "all done"

