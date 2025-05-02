# """obliv: ORAM with variable-size blocks and HIRB data structure."""

# from setuptools import setup, find_packages
# from os import path

# here = path.abspath(path.dirname(__file__))

# with open(path.join(here, 'README.rst')) as f:
#     long_description = f.read()


# setup(
#     name='obliv',
#     version='0.0.2',
#     description='ORAM with variable-size blocks and HIRB data structure',
#     long_description=long_description,
#     author='Daniel S. Roche',
#     author_email='roche@usna.edu',
#     license='Unlicense',
#     packages=find_packages(),
#     install_requires=[
#         'pycryptodome>=3.18.0',
#         'paramiko>=3.0.0'
#     ],
#     extras_require={
#         'dev': [
#             'pytest>=7.0.0',
#             'coverage>=6.0.0',
#             'mypy>=1.0.0',
#             'flake8>=5.0.0',
#         ]
#     },
#     tests_require=[
#         'pycryptodome>=3.18.0',
#         'paramiko>=3.0.0'
#     ],
#     test_suite='tests',
#     python_requires='>=3.6, <3.14',
#     classifiers=[
#         'Programming Language :: Python :: 3',
#         'Operating System :: OS Independent'
#     ]
#     # packages=['obliv'],
#     # install_requires=['pycrypto', 'paramiko'],
# )
