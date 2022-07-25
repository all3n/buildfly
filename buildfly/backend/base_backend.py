class BaseBackend(object):
    def __init__(self, ctx):
        self.name = "base"
        self.ctx = ctx

    def setup(self):
        pass

    def generate(self):
        pass

    def build(self):
        pass