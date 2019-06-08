from buildfly.generator.base_generator import base_generator
# from buildfly.targets.cc_targets import CcTarget



class makefile_generator(base_generator):
    def write(self, target):
        BUILD_RULE="{target}: ${SRCS_OBJ}\n".format(
            target=target.name,
            SRCS_OBJ=",".join(target.srcs)
        )
        self.write_line(BUILD_RULE)

