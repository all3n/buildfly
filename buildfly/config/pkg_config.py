from collections import defaultdict


class PkgConfigFile(object):

    def replace_with_vars_map(self, string, var_map):
        for key, value in var_map.items():
            string = string.replace("${" + key + "}", str(value))
        return string

    def __init__(self, p):
        self.path = p
        self.props = defaultdict(str)
        with open(self.path, "r") as f:
            lines = f.readlines()
            for line in lines:
                if line == '\n':
                    continue
                    description = line.rstrip()
                elif "=" in line:
                    name, val = line.rstrip().split("=")
                    self.props[name] = self.replace_with_vars_map(val, self.props)
                elif ": " in line:
                    name, val = line.rstrip().split(": ")
                    self.props[name.lower()] = self.replace_with_vars_map(val, self.props)

    def has(self, name):
        return name in self.props

    def get(self, name, dv=None):
        return self.props.get(name, dv)

    def set(self, name, dv):
        self.props[name] = dv

# x = PkgConfigFile("/usr/lib/x86_64-linux-gnu/pkgconfig/zlib.pc")
# print(x.props)
