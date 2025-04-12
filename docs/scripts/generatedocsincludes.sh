#!/bin/bash


set -eu

mydir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

outdir="$mydir/../_static/includes/"


# $1 is output file name without a path
# $2 is the command line
function generatedocinsert() {
    outfile=$outdir$1
    cmd="$2"

 #   echo ".. code:: bash"> $outfile
 #   echo ""  >> $outfile
    echo "[..botcoin] \$ $cmd"  > $outfile
    $cmd | sed -e"s?$HOME?/home/user?g" >> $outfile
}


generatedocinsert botcoin_version.txt "botcoin version"
generatedocinsert botcoin_help.txt "botcoin help"
generatedocinsert botcoin_keys_help.txt "botcoin keys help"
generatedocinsert botcoin_help_keys_new.txt "botcoin help keys new"
generatedocinsert botcoin_help_keys_inspect.txt "botcoin help keys inspect"
generatedocinsert botcoin_help_keys_update.txt "botcoin help keys update"
generatedocinsert botcoin_help_keys_list.txt "botcoin help keys list"
generatedocinsert botcoin_help_config_location.txt "botcoin help config location"
generatedocinsert botcoin_help_config_build.txt "botcoin help config build"
generatedocinsert botcoin_help_config_pull.txt "botcoin help config pull"
generatedocinsert botcoin_help_run.txt "botcoin help run"


generatedocinsert giverny_help_network_new.txt "giverny help network new"
generatedocinsert giverny_version.txt "giverny version"
generatedocinsert giverny_help_keys_import.txt "giverny help keys import"

generatedocinsert giverny_help_keys_generate.txt "giverny help keys generate"
generatedocinsert giverny_help_network_push.txt "giverny help network push"

generatedocinsert giverny_help_network_add.txt "giverny help network add"
generatedocinsert giverny_help_network_start.txt "giverny help network start"
generatedocinsert giverny_help_network_stop.txt "giverny help network stop"
generatedocinsert giverny_help_transactions_solo.txt "giverny help transactions solo"


generatedocinsert giverny_parse.txt "giverny help parse" 
# Special cases

giverny help keys | grep -A10 "Global Flags:" > $outdir"giverny_keys_flags.txt"


