document.querySelectorAll("div.post").forEach(
    (e) => e.querySelectorAll("a.r").forEach(
	(q) => ((id, hlr) => Object.keys({mouseout: false, mouseover: true}).map(
	    (ev, hl) => q.addEventListener(ev, hlr(id, hl))
	))(q.href.replace(/^.*#/ ,""),
	   (id, hl) => (() =>document.getElementById(id).style.background = hl ? "#ccd" : ""))));
