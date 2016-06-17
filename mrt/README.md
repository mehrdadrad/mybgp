# Export Multi-Threaded Routing Toolkit (MRT) Routing Information

- JSON file
- MongoDB

```
#go run mrt.go -help
  -format string
    	export format (default "json")
  -jsonfile string
    	export file full path (default "export_mrt.json")
  -mrtfile string
    	enter the full MRT path
    	
#go run mrt.go -mrtfile /tmp/rib.20160616.1600 -jsonfile rib.json    	
```    	
