function openList() {
	$("#listBtn").addClass("active");
	$("#compareBtn").removeClass("active");

	$("#compareDiv").hide();
	$("#listDiv").show();
}

function openCompare() {
	$("#listBtn").removeClass("active");
	$("#compareBtn").addClass("active");

	$("#listDiv").hide();
	$("#compareDiv").show();
}

$("#listBtn").click(openList)
$("#compareBtn").click(openCompare)

var hash = window.location.hash.substring(1);
if (hash == "compare")
	openCompare();
else
	openList();
