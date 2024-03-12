# axfrcheck
Simple tool to check if a master can give a proper axfr to a slave, using the named.conf 
## ToDo
 - Use viper for cli configuration.
 - Detect named or pdns config on the fly and parse accordingly.
 - Parametrize number of workers.
 - Assumes that the slave is ok if only 1 master is ok, make that configurable too.
