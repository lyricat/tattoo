#TATTOO!

##Demo

 - My Blog: [shellex.info](http://shellex.info)

##Build & Install & Run

### requirements

 - [Go weekly](http://weekly.golang.org)
 - [blackfriday](https://github.com/shellex/blackfriday) that compatible with go weekly. ( orig: [russross](https://github.com/russross/blackfriday) )
		
### build & install
 
use follow command to install blackfriday from my repo

	go get github.com/shellex/blackfriday
	go install github.com/shellex/blackfriday

build and install tattoo

	./all.sh

### as a stand-alone server

	./tattoo

### with Fast-CGI

#### configure nginx (Fast-CGI)

add the following code to your nginx site config file to make it work:

	location /static/ {
		expires 1d;
		root YOUR_THEME_PATH;
		add_header Cache-Control  must-revalidate;
	}
	location /sys/static/ {
		expires 1d;
		root YOUR_BLOG_PATH;
		add_header Cache-Control  must-revalidate;
	}
	location / {
		root  YOUR_BLOG_PATH;
		expires 5m;
		add_header Cache-Control  must-revalidate;
		include fastcgi_params;
		fastcgi_param REQUEST_METHOD $request_method;
		fastcgi_param QUERY_STRING $query_string;
		fastcgi_param CONTENT_TYPE $content_type;
		fastcgi_param CONTENT_LENGTH $content_length;
		fastcgi_param GATEWAY_INTERFACE CGI/1.1;
		fastcgi_param SERVER_SOFTWARE nginx/$nginx_version;
		fastcgi_param REMOTE_ADDR $remote_addr;
		fastcgi_param REMOTE_PORT $remote_port;
		fastcgi_param SERVER_ADDR $server_addr;
		fastcgi_param SERVER_PORT $server_port;
		fastcgi_param SERVER_NAME $server_name;
		fastcgi_param SERVER_PROTOCOL $server_protocol;
		fastcgi_param SCRIPT_FILENAME $fastcgi_script_name;
		fastcgi_param PATH_INFO $fastcgi_script_name;
		fastcgi_pass 127.0.0.1:8887;
	}

and then, run

	./tattoo -fcgi



