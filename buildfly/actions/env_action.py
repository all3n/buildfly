#!/usr/bin/env python
# -*- coding: utf-8 -*-
# Author wanghuacheng
# Description
# Date 10/11/22 3:20 PM
# Version 1.0
import json
import os.path

from buildfly.actions.base_action import BaseAction
from buildfly.actions.get_action import GetAction
from buildfly.utils.io_utils import write_to_file
from buildfly.utils.log_utils import get_logger
import yaml

from buildfly.utils.system_utils import get_bfly_path

logger = get_logger(__name__)

ACTIVATE_TMP = """
# This file must be used with "source bin/activate" *from bash*
# you cannot run it directly

bin=`dirname "$0"`
export ENV_HOME=`cd "$bin/../"; pwd`


if [ "${BASH_SOURCE-}" = "$0" ]; then
    echo "You must source this script: \$ source $0" >&2
    exit 33
fi

bdeactivate () {
    # reset old environment variables
    # ! [ -z ${VAR+_} ] returns true if VAR is declared at all
    
    if ! [ -z "${_OLD_B_PATH:+_}" ] ; then
        PATH="$_OLD_B_PATH"
        export PATH
        unset _OLD_B_PATH
    fi
    
    if ! [ -z "${_OLD_LD_LIBRARY_PATH:+_}" ] ; then
        LD_LIBRARY_PATH=$_OLD_LD_LIBRARY_PATH
        export LD_LIBRARY_PATH
        unset _OLD_LD_LIBRARY_PATH
    fi
    
    if ! [ -z "${_OLD_LIBRARY_PATH:+_}" ] ; then
        LD_LIBRARY_PATH=$_OLD_LIBRARY_PATH
        export LIBRARY_PATH
        unset _OLD_LIBRARY_PATH
    fi
    if ! [ -z "${_OLD_C_INCLUDE_PATH:+_}" ] ; then
        C_INCLUDE_PATH=$_OLD_C_INCLUDE_PATH
        export C_INCLUDE_PATH
        unset _OLD_C_INCLUDE_PATH
    fi
    
    if ! [ -z "${_OLD_CPLUS_INCLUDE_PATH:+_}" ] ; then
        CPLUS_INCLUDE_PATH=$_OLD_CPLUS_INCLUDE_PATH
        export CPLUS_INCLUDE_PATH
        unset _OLD_CPLUS_INCLUDE_PATH
    fi
    
    if ! [ -z "${_OLD_VIRTUAL_BENV_HOME+_}" ] ; then
        BENV_HOME="$_OLD_VIRTUAL_BENV_HOME"
        export BENV_HOME
        unset _OLD_VIRTUAL_BENV_HOME
    fi

    # The hash command must be called to get it to forget past
    # commands. Without forgetting past commands the $PATH changes
    # we made may not be respected
    hash -r 2>/dev/null

    if ! [ -z "${_OLD_VIRTUAL_PS1+_}" ] ; then
        PS1="$_OLD_VIRTUAL_PS1"
        export PS1
        unset _OLD_VIRTUAL_PS1
    fi

    unset BENV
    unset BENV_NAME
    unset BENV_CFG
    if [ ! "${1-}" = "nondestructive" ] ; then
    # Self destruct!
        unset -f bdeactivate
    fi
}
# unset irrelevant variables
bdeactivate nondestructive

export BENV_NAME={benv}
export BENV_CFG={benv_cfg}
BENV=$ENV_HOME
if ([ "$OSTYPE" = "cygwin" ] || [ "$OSTYPE" = "msys" ]) && $(command -v cygpath &> /dev/null) ; then
    BENV=$(cygpath -u "$BENV")
fi
export BENV

_OLD_B_PATH="$PATH"
PATH="$BENV/bin:$PATH"
_OLD_LD_LIBRARY_PATH=$LD_LIBRARY_PATH
LD_LIBRARY_PATH=$BENV/lib:$LD_LIBRARY_PATH
export PATH
export LD_LIBRARY_PATH
export _OLD_LD_LIBRARY_PATH


_OLD_LIBRARY_PATH=$LIBRARY_PATH
export _OLD_LIBRARY_PATH
export LIBRARY_PATH=$BENV/lib:$LIBRARY_PATH

_OLD_C_INCLUDE_PATH=$C_INCLUDE_PATH
export _OLD_C_INCLUDE_PATH
export C_INCLUDE_PATH=$BENV/include:$C_INCLUDE_PATH

_OLD_CPLUS_INCLUDE_PATH=$CPLUS_INCLUDE_PATH
export _OLD_CPLUS_INCLUDE_PATH
export CPLUS_INCLUDE_PATH=$BENV/include:$CPLUS_INCLUDE_PATH


# unset BENV_HOME if set
if ! [ -z "${BENV_HOME+_}" ] ; then
    _OLD_VIRTUAL_BENV_HOME="$BENV_HOME"
    unset BENV_HOME
fi

if [ -z "${BENV_DISABLE_PROMPT-}" ] ; then
    _OLD_VIRTUAL_PS1="${PS1-}"
    if [ "x" != x ] ; then
        PS1="${PS1-}"
    else
        PS1="\033[30;46;01m(BENV:$BENV_NAME)\033[0m ${PS1-}"
    fi
    export PS1
fi


# The hash command must be called to get it to forget past
# commands. Without forgetting past commands the $PATH changes
# we made may not be respected
hash -r 2>/dev/null

"""

#https://gcc.gnu.org/onlinedocs/gcc/Environment-Variables.html

class EnvAction(BaseAction):
    def parse_args(self, parser):
        parser.add_argument('-f', metavar='config file', type=str,
                            help="config file")
        parser.add_argument('env', metavar='env', type=str,
                            help="env name")

        parser.add_argument('args', metavar='args', type=str, nargs="*",
                            help="cmd args")

    def run(self):
        env = self.args.env
        args, kwargs = self.split_args()
        benv = kwargs.get("name", "benv")
        if env == "update":
            if 'BENV_NAME' in os.environ and 'BENV' in os.environ:
                benv = kwargs.get("name", os.environ['BENV_NAME'])
                env = kwargs.get("env_path", os.environ['BENV'])
                env_cfg = kwargs.get("cfg", os.environ['BENV_CFG'])
                ac_file = os.path.join(env, 'bin', 'activate')
                write_to_file(ac_file, ACTIVATE_TMP.replace("{benv}", benv).replace( "{benv_cfg}", env_cfg))
            else:
                logger.error("not in benv,activate first")
                return
        else:
            if not os.path.exists(env):
                os.makedirs(env)
                os.makedirs(os.path.join(env, "bin"))
                os.makedirs(os.path.join(env, "include"))
                os.makedirs(os.path.join(env, "lib"))

                ac_file = os.path.join(env, 'bin', 'activate')
                write_to_file(ac_file, ACTIVATE_TMP.replace("{benv}", benv).replace("{benv_cfg}", os.path.abspath(os.path.realpath(self.args.f))))
                os.chmod(ac_file, 0o664)
        benv_json = os.path.join(env, "benv.json")
        cfg_yaml = os.environ.get('BENV_CFG', self.args.f)
        if not cfg_yaml:
            logger.error(f"cfg_yaml not set")
            return
        with open(cfg_yaml, "rb") as f:
            env_cfg = yaml.load(f.read(), yaml.FullLoader)
        deps = env_cfg['deps']
        get_action = GetAction()
        envs_files = {}
        for idx, dep in enumerate(deps):
            dep_files = []
            logger.info("[%d] %s", idx, dep)
            if type(dep) == str:
                d_args = []
                d_kwargs = {}
                d_name = dep
            elif type(dep) == dict:
                d_args = []
                d_kwargs = dep
                d_name = dep['name']
                # print(dep)
            else:
                raise "dep %s type not support" % (dep)

            bpkg = get_action.get_pkg(d_name, d_args, d_kwargs)

            artifact_dir = get_bfly_path(f'install/{bpkg.artifact_path}')
            print(artifact_dir)
            dirs = ['bin', 'include', 'lib']

            for d in dirs:
                env_dir = os.path.join(env, d)
                dpath = os.path.join(artifact_dir, d)
                if not os.path.exists(dpath):
                    continue
                for sd in os.listdir(dpath):
                    sub_path = os.path.join(dpath, sd)
                    if sd in ['cmake', 'pkgconfig']:
                        continue
                    target_env_dir = os.path.join(env_dir, sd)
                    dep_files.append(os.path.join(d, sd))
                    if os.path.exists(target_env_dir):
                        continue
                    os.symlink(sub_path, target_env_dir)
            envs_files[dep] = dep_files
        benv_info = {
            'deps_files': envs_files
        }
        write_to_file(benv_json, json.dumps(benv_info, indent=2))