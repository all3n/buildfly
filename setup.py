#!/usr/bin/env python
from setuptools import setup, find_packages
import os
import pathlib


def search_file(d, base):
    ret = []
    dret = []
    for sd in os.listdir(d):
        sdf = os.path.join(d, sd)
        if os.path.isfile(sdf):
            dret.append(sdf)
        else:
            ret.extend(search_file(sdf, base + "/" + sd))
    ret.append((base, dret))
    return ret


here = pathlib.Path(__file__).parent.resolve()
long_description = (here / "README.md").read_text(encoding="utf-8")
setup(name='buildfly',
      version='1.0',
      description='build c++ fly',
      author='all3n',
      author_email='wanghch8398@163.com',
      url='https://github.com/all3n/buildfly',
      long_description=long_description,
      long_description_content_type="text/markdown",
      classifiers=[  # Optional
          # How mature is this project? Common values are
          #   3 - Alpha
          #   4 - Beta
          #   5 - Production/Stable
          "Development Status :: 3 - Alpha",
          # Indicate who your project is intended for
          "Intended Audience :: Developers",
          "Topic :: Software Development :: Build Tools",
          # Pick your license as you wish
          "License :: OSI Approved :: MIT License",
          # Specify the Python versions you support here. In particular, ensure
          # that you indicate you support Python 3. These classifiers are *not*
          # checked by 'pip install'. See instead 'python_requires' below.
          "Programming Language :: Python :: 3",
          "Programming Language :: Python :: 3.7",
          "Programming Language :: Python :: 3.8",
          "Programming Language :: Python :: 3.9",
          "Programming Language :: Python :: 3.10",
          "Programming Language :: Python :: 3 :: Only",
      ],
      install_requires=[
          "six",
          "pyyaml",
          "colorama",
          "requests",
          "semver",
          "sqlalchemy"
      ],
      entry_points={
          'console_scripts': [
              'bfly=buildfly.__main__:main'
          ]
      },
      packages=find_packages("."),
      include_package_data=True
      )
