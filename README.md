##Betarigs autoprice (brAutoprice)

brAutoprice is a tool that helps you to rent your rig at the best price.

If market prices go up, your rental price goes up. So you always rent at the best price.

If market prices go down, your rental price goes down BUT never be lower than your defined min price. So you are always on the top of the list without the risk to rent under your rentability price. 

If this tool is useful for you and/or if it helps you to win more money, please consider to make a donation. Donations will be used to improve this tool and work on other ones.

![Donation QR](http://api.qrserver.com/v1/create-qr-code/?size=200x200&data=bitcoin:1HgpsmxV52eAjDcoNpVGpYEhGfgN7mM1JB%3Flabel%3DToorop)

[1HgpsmxV52eAjDcoNpVGpYEhGfgN7mM1JB](http://tinyurl.com/mccsoez)


###Download Binaries
Coming soon

* Linux: 
* MacOs: 
* Windows:

###Options

* --apiKey (required): your Betarigs API key
* -- minPrice (required): the min price per speed unit for your rig. brAutoprice will never set rental price below this limit.
* --rigId (required): the ID of the concerned rig  
* --priceDiff: the diff in percent between the lower price and the price you want for your rig

####Examples


	./brAutoprice --apiKey XXXXX --rigId 4568 --minPrice 0.0008 --priceDiff -1
	
In this example your rig will always be on the fisrt position if the market price is upper than 0.0008.
If the market price goes UP, your rig price goes up but you rig still stay in first position
	
	