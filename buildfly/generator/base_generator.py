

class base_generator(object):
    def __init__(self, gen_file):
        self.gen_file = open(gen_file, "w")

    def __del__(self):
        self.close()

    def write_line(self, content):
        self.gen_file.write(content)


    def close(self):
        if self.gen_file:
            self.gen_file.close()
            self.gen_file = None

