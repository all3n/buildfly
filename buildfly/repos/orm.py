#!/usr/bin/env python
# -*- coding: utf-8 -*-
# Author wanghuacheng
# Description
# Date 10/11/22 2:15 PM
# Version 1.0
import logging

from sqlalchemy import Column, TEXT
from sqlalchemy import ForeignKey
from sqlalchemy import Integer
from sqlalchemy import String, TIMESTAMP
from sqlalchemy.orm import declarative_base
from sqlalchemy.orm import relationship

Base = declarative_base()

class Repos(Base):
    __tablename__ = "repos"
    name = Column(String(50), primary_key=True)
    path = Column(String)
    desc = Column(String)
    ts = Column(TIMESTAMP)

    def __repr__(self):
        return f"Repos(name={self.name!r}, path={self.path!r}, desc={self.desc!r}, ts={self.ts!r})"


class Pkgs(Base):
    __tablename__ = "pkgs"
    id = Column(Integer, primary_key=True)
    name = Column(String(50))
    repo = Column(String(50))
    path = Column(String)
    desc = Column(String)
    ts = Column(TIMESTAMP)

    def __repr__(self):
        return f"Pkgs(id={self.id!r}, name={self.name!r}, path={self.path!r}, desc={self.desc!r}, ts={self.ts!r})"


class Artifact(Base):
    __tablename__ = "artifact"
    id = Column(Integer, primary_key=True)
    pkg = Column(String(50))
    ver_id = Column(String(100))
    build_params = Column(TEXT)
    param_hash = Column(String(50))
    hash = Column(String(50))
    os = Column(String)
    arch = Column(String)
    ts = Column(TIMESTAMP)

    def __repr__(self):
        return f"Artifact(id={self.id!r}, name={self.name!r}, ts={self.ts!r})"
