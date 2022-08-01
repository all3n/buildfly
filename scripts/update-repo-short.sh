bin=`dirname "$0"`
export SRC_HOME=`cd "$bin/.."; pwd`

curl https://raw.githubusercontent.com/fffaraz/awesome-cpp/master/README.md|grep -P -o "https://github.com/[-_a-zA-Z0-9]+/[-_a-zA-Z0-9]+" \
 |sort|uniq|awk -F '/' 'BEGIN{printf("repo_list = {\n")}{printf("\t\"%s\": \"%s/%s\",\n", tolower($NF),$4,$5)}END{printf("}")}' \
  > ${SRC_HOME}/buildfly/config/repo_list.py
