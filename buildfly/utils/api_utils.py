import threading
from functools import wraps

import buildfly.api

context = threading.local()


def get_build_instance():
    return getattr(context, 'buildfly_build', None)


def set_build_instance(build_instance):
    context.buildfly_build = build_instance


class BuildFlyAPI(object):
    def __init__(self, build_instance):
        self.build_instance = build_instance

    def __enter__(self):
        self.old_build_instance = get_build_instance()
        set_build_instance(self.build_instance)

    def __exit__(self, _type, _value, _tb):
        set_build_instance(self.old_build_instance)


def bfly_api_method(f):
    @wraps(f)
    def wrapped(*args, **kwargs):
        build_instance = get_build_instance()
        if build_instance is None:
            raise RuntimeError(
                'buildfly api method  %s must call in build stage'
                % f.__name__
            )
        return getattr(build_instance, f.__name__)(*args, **kwargs)

    setattr(buildfly.api, f.__name__, wrapped)
    buildfly.api.__all__.append(f.__name__)
    f.is_api_method = True
    return f
