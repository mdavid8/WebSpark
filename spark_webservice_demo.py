# Copyright 2015 David Wang. All rights reserved.
# Use of this source code is governed by MIT license.
# Please see LICENSE file

# WebSpark 
# Spark web service demo

# version 0.2

# use REPL or define sc SparkContext 

import urllib2, urllib
import math
import demohtml
import time
import traceback


# Spark Web Application demo with parallel processing
# see demoservice function

ServerAddr="http://<enter WebSpark IP address here>:8001"
RegisterURL=ServerAddr + "/addapi?"
RespondURL=ServerAddr + "/respond?"
errwaitseconds = 3

element = '<li class="list-group-item">first prime above %d is %d</li>'


with open('template.html') as f:
	template = f.read()

def slow_isprime(num):
	if num<2:
		return False
	for i in range(2, int(math.sqrt(num))+1):
		if num%i==0:
			return False
	return True


def firstprimeabove(num):
	i=num+1
	while True:
		if slow_isprime(i):
			return i
		i+=1 


servicename = 'demo'


# Spark Web Application demo
def demo(url):
	rawdata = range(1000, 20000, 1100)
	data = sc.parallelize(rawdata)
	above=data.map(lambda x: (x, firstprimeabove(x))).collect()
	primelist=[element%x for x in above]
	response = template % ' '.join(primelist)
	return response


def parserequest(rawrequest):
	lines = rawrequest.split('\n')
	if len(lines)<4:
		print 'incorrect WebSpark request'
	else:
		name = lines[0]
		url = lines[1]
		remoteaddr = lines[2]
		header = lines[3:]
		return name, url, remoteaddr, header


st =''

# publish web service with WebSpark
while True:
	try:
		url = RegisterURL + urllib.urlencode({'name': servicename})
		conn = urllib2.urlopen(url)
		data = conn.read()
		conn.close()
		name, clienturl, remoteaddr, header = parserequest(data)
		print name, clienturl, remoteaddr, header
		response = demo(clienturl)
		url = RespondURL + urllib.urlencode({'name': name})
		conn = urllib2.urlopen(url, response)
		conn.close()
	except Exception as ex:
		print 'error connecting to WebSpark at', ServerAddr
		traceback.print_exc()
		time.sleep(errwaitseconds)
		continue


