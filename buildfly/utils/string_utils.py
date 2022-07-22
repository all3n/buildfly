import re


def underscore(word):
    """
    Make an underscored, lowercase form from the expression in the string.

    Example::

        >>> underscore("DeviceType")
        "device_type"

    As a rule of thumb you can think of :func:`underscore` as the inverse of
    :func:`camelize`, though there are cases where that does not hold::

        >>> camelize(underscore("IOError"))
        "IoError"

    """
    word = re.sub(r"([A-Z]+)([A-Z][a-z])", r'\1_\2', word)
    word = re.sub(r"([a-z\d])([A-Z])", r'\1_\2', word)
    word = word.replace("-", "_")
    return word.lower()


def camelize(string, uppercase_first_letter=True):
    """
    Convert strings to CamelCase.

    Examples::

        >>> camelize("device_type")
        "DeviceType"
        >>> camelize("device_type", False)
        "deviceType"

    :func:`camelize` can be thought of as a inverse of :func:`underscore`,
    although there are some cases where that does not hold::

        >>> camelize(underscore("IOError"))
        "IoError"

    :param uppercase_first_letter: if set to `True` :func:`camelize` converts
        strings to UpperCamelCase. If set to `False` :func:`camelize` produces
        lowerCamelCase. Defaults to `True`.
    """
    if uppercase_first_letter:
        return re.sub(r"(?:^|_)(.)", lambda m: m.group(1).upper(), string)
    else:
        return string[0].lower() + camelize(string)[1:]


def replace_with_vars_map(string, var_map):
    for key, value in var_map.items():
        string = string.replace("{" + key + "}", str(value))
    return string


def is_int(sval: str):
    try:
        int(sval)
    except:
        return False
    else:
        return True


def is_num(sval: str):
    """
    because str.isnumeric can't detact negative number
    use exception to detact
    :param sval:str
    :return:
    """
    try:
        float(sval)
    except:
        return False
    else:
        return True
