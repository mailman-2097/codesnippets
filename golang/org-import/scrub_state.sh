#!/bin/bash
function clean_tf_state () {
  scrub_list=($(terraform state list))
  echo "Current tf state has ${#scrub_list[@]} items to be scrubbed"
  echo "${scrub_list[@]}"
  echo "---"
  for ((i = 0; i < "${#scrub_list[@]}"; i++)); do
    terraform state rm "${scrub_list[i]}"
    echo "---"
  done
}
read -t 10 -n 1 -p "Are you sure (Yes/No)? " answer
case ${answer:0:1} in
    y|Y )
        clean_tf_state
    ;;
    * )
        echo "Abort"
    ;;
esac
