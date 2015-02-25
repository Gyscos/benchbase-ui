function getBenchList(host, filter, depth, focus) {
	$.getJSON(host + "/list?filter=" + filter, function(data) {
		if (data["Error"] != null) {
			// Error?
			alert(data["Error"]);
			return;
		}

		// r is now a list of benchmarks
		var benchlist = data["Result"];

		if (focus != "") {
			cutOutOfFocus(benchlist, focus);
			if (depth != 0) {
				var focusList = focus.split(".");
				depth += focusList.length;
			}
		}

		if (depth != 0)
			truncateDeepResults(benchlist, depth);

		// Group then by compatible configuration?
		var groups = groupByConfiguration(benchlist);

		// Now for each group, show a table
		$.each(groups, function(i, group) {
			// group is a list of benchmarks
			var t = makeTable(group, filter);
			$("#list").append(t);
		});
	});
}

function cutOutOfFocus(benchlist, focus) {
	$.each(benchlist, function(i, bench) {
		var r = bench["Result"];
		var newResult = {};
		$.each(r, function(key, value) {
			if (key.substring(0, focus.length) == focus) {
				newResult[key] = value;
			}
		});
		bench["Result"] = newResult;
	});
}

function truncateDeepResults(benchlist, depth) {
	$.each(benchlist, function(i, bench) {
		var r = bench["Result"];
		var newResult = {};
		$.each(r, function(key, value) {
			var l = key.split(".");
			if (l.length <= depth) {
				newResult[key] = value;
			} else if (l.length == depth+1) {
				if (l[l.length-1] == "total") {
					newResult[l.slice(0,-1).join(".")] = value;
				}
			}
		});
		bench["Result"] = newResult;
	});
}

function groupByConfiguration(benchlist) {

	var result = [];

	var map = {"true":[], "false":[]};

	$.each(benchlist, function(i, data) {
		map[data["Conf"]["ForceAnalyze"]].push(data);
	});

	$.each(map, function(key, value) {
		if (value.length > 0)
			result.push(value);
	});

	return result;
}

function makeTitle(configuration) {
	if (configuration["ForceAnalyze"])
		return "Analyze API";
	else
		return "Direct API";
}

function makeTable(benchlist, filter) {
	var t = $("<table>");
	t.addClass("table");
	t.addClass("table-hover");
	t.addClass("table-striped");
	t.addClass("table-bordered");
	t.addClass("table-nonfluid");

	if (benchlist.length == 0)
		return t;

	var head = $("<thead>");
	t.append(head);


	var m = makeResultTree(benchlist[0]["Result"]);
	mergeSingleChilds(m);
	addTimeLabels(head, m, benchlist[0]["Conf"], filter);

	var body = $("<tbody>");
	t.append(body);

	$.each(benchlist, function(i, bench) {
		var bodyTr = $("<tr>");
		body.append(bodyTr);

		// Entry label: the configuration
		bodyTr.append("<th scope='row'>" + getReadableConf(bench["Conf"]) + "</th>");
		// Entry values...
		addTimeResults(bodyTr, m, "", bench["Result"]);
	});

	return t;
}

function addTimeLabels(head, resultTree, configuration, filter) {
	getWidthDepth(resultTree);

	var title = makeTitle(configuration);
	var titleTr = $("<tr>");
	head.append(titleTr);
	var titleTh = $("<th>");
	titleTr.append(titleTh);
	titleTh.attr("colspan", 1+resultTree.dw.width);
	titleTh.html(title);

	var maxDepth = resultTree.dw.depth;

	var heap = [];
	$.each(resultTree.children, function(i,child) {
		heap.push(child);
	});

	var depth = 0;
	while (heap.length > 0) {
		depth++;
		var headTr = $("<tr>");
		head.append(headTr);
		newHeap = [];

		if (depth == 1) {
			var th = $("<th>");
			th.html("Host");
			th.attr("rowspan", maxDepth);
			headTr.append(th);
		}

		$.each(heap, function(i, data) {
			// Handle this depth
			var th = $("<th>");
			headTr.append(th);
			data["th"] = th;
			var focus = ("prefix" in data) ? data.prefix : "";
			focus += data.name;
			th.html('<a href="/list?focus=' + focus + '&depth=1&filter=' + filter + '">' + data.name + "<a>");
			th.attr('colspan', data.dw.width);
			if (data.dw.depth == 1)
				th.attr('rowspan', maxDepth - depth);

			$.each(data.children, function(i, child) {
				newHeap.push(child);
			});
		});

		heap = newHeap;
	}
}

function getWidthDepth(node) {
	var maxDepth = 0;
	var width = 0;
	if (node.children.length == 0)
		width = 1;
	$.each(node.children, function(i, child) {
		var dw = {};
		if ("dw" in child) {
			dw = child.dw;
		} else {
			dw = getWidthDepth(child);
		}
		width += dw.width;
		maxDepth = Math.max(maxDepth, dw.depth);
	});
	var dw = {"width":width, "depth":maxDepth+1};
	node["dw"] = dw;
	return dw;
}

function addTimeResults(body, resultTree, prefix, result) {
	$.each(resultTree.children, function(i, child) {
		if (child.children.length == 0) {
			if (child.name == "total")
				element = $("<th>");
			else
				element = $("<td>");
			element.attr("align", "right");
			element.html(result[prefix+child.name].toFixed(3));
			body.append(element);
		} else {
			addTimeResults(body, child, prefix + child.name + ".", result);
		}
	});
}

function getReadableConf(configuration) {
	var result = configuration["Host"];
	if (isNormalInteger(configuration["Rev"]))
		result += " r" + configuration["Rev"];
	result += " (" + configuration["Threads"] + "&nbsp;threads)";

	return result;
}

function isNormalInteger(str) {
    return /^\+?(0|[1-9]\d*)$/.test(str);
}

function findChild(node, childName) {
	var result = -1;
	$.each(node.children, function(i, child) {
		if (child.name == childName) {
			result = i;
			return false;
		}
	});
	return result;
}

function mergeSingleChilds(tree) {
	if ("name" in tree && tree.name != "" && tree.children.length == 1) {
		tree.name += "." + tree.children[0].name
		tree.children = tree.children[0].children;
		mergeSingleChilds(tree);
	} else {
		$.each(tree.children, function(i, child) {
			mergeSingleChilds(child);
		});
	}
}

function makeResultTree(result) {
	var m = {"children":[]};

	$.each(result, function(key, value) {
		var l = key.split('.');
		var current = m;
		$.each(l, function(i, token) {
			var i = findChild(current, token);
			if (i == -1) {
				child = {"name":token,"children":[]};
				if ("name" in current)
					if ("prefix" in current)
						child["prefix"] = current.prefix + current.name + ".";
					else
						child["prefix"] = current.name + ".";
				if (child.name == "total")
					current.children.unshift(child);
				else
					current.children.push(child);
				i = current.children.length - 1;
			}
			current = current.children[i];
		});
	});

	return m;
}
