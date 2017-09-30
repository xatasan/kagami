
var error = document.getElementById("error");
var list = document.getElementById("list");

// from https://stackoverflow.com/questions/901115
function getParameterByName(name, url) {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&");
    var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}

function genPost(p) {
    var post = document.createElement("div");
    post.setAttribute("class", "post resp");
    post.setAttribute("id", p.postno);

    //// HEADER //////////////////////////////////////
    var header = document.createElement("header");
    if (p.subject) {
	var subject = document.createElement("span");
	subject.setAttribute("class", "subject");
	subject.innerHTML = p.subject;
	header.appendChild(subject);
    }
	
    var name = document.createElement("span");
    name.setAttribute("class", "name");
    name.innerHTML = p.name;
    header.appendChild(name);

    var time = document.createElement("time");
    time.innerHTML = p.time;
    header.appendChild(time);

    if (p.capcode) {
	var capcode = document.createElement("span");
	capcode.setAttribute("class", "capcode");
	capcode.innerHTML = "##" + p.capcode;
	header.appendChild(capcode);
    }

    var plink = document.createElement("a");
    plink.setAttribute("class", "ref");
    plink.setAttribute("href", "./res/"+p.respto+".html#"+p.postno);
    plink.innerHTML = "No. ";
    header.appendChild(plink);
    header.innerHTML += p.postno;
    
    post.appendChild(header);

    //// ASIDE ///////////////////////////////////////
    var aside = document.createElement("aside");
    aside.setAttribute("class", "atts");
    
    post.appendChild(aside);

    //// MAIN ////////////////////////////////////////
    var main = document.createElement("main");
    main.innerHTML = p.comment;
    post.appendChild(main);
    
    return post;
}

function displayResults(data) {
    while (list.hasChildNodes())
	list.removeChild(list.lastChild);

    data.forEach((post) => ((li) => {
	li.appendChild(genPost(post));
	list.appendChild(li);
    })(document.createElement("li")));
}

function search(query, limit, page) {
    error.style.display = "none";
    var request = new XMLHttpRequest();
    request.open('GET', "/search"+ // has to be set on a per-server basis!
		 "?q="+query+
		 "&p="+(page||1)+
		 "&l="+(limit||25), true);
    request.onerror = () => {
	error.style.display = "block";
	error.innerHTML = "There was an error";
    };
    request.onload = () => (request.status >= 200 && request.status < 400) ?
	displayResults(JSON.parse(request.responseText)) :
	error.innerHTML = "There was an error while trying to reach the server";
    request.send();
}

window.onload = () => ((q) => q ? search(q, 25, 1) : 0)(getParameterByName("q"));
