import os
class base_generator(object):
    def __init__(self, src_dir,build_dir):
        self.src_dir = src_dir
        self.build_dir = build_dir
        self.gen_file = open(os.path.join(build_dir, "Makefile"), "w")
        self.write_common()

    def write_common(self):
        pass

    def __del__(self):
        self.close()

    def write_line(self, content):
        self.gen_file.write(content)

    def close(self):
        if self.gen_file:
            self.gen_file.close()
            self.gen_file = None
