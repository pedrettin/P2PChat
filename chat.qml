import QtQuick 2.0

 Rectangle {
     id: window
     width: 800; height: 600
     color: "lightgray"
     //bottom entry line
     Rectangle {
		id: entryRect
		anchors.bottom: parent.bottom
		anchors.left: parent.left
		width: parent.width
		height: 50
		border.color: "black"
		TextInput {
	    	id: entry
	    	anchors.fill: parent
	    	Keys.onReturnPressed: { ctrl.textEntered(entry); text=""}
	    	Keys.onEnterPressed: { ctrl.textEntered(entry); text=""}
	    	
	    	focus: true
	    	wrapMode: "WordWrap"
		}
     }
     //user list
     Rectangle {
		id: listRect
		anchors.top: parent.top
		anchors.right: parent.right
		width: 200
		height: parent.height-entryRect.height
		border.color: "black"
		Text {
		id: userlist
		objectName: "userlist"
	    	text: ""
	    	anchors.fill: parent
	    	wrapMode: "WordWrap"
		}
     }
     //conversation area
     Rectangle {
		id: convRect
		anchors.top: parent.top
		anchors.left: parent.left
		width: parent.width - listRect.width
		height: parent.height - entryRect.height
		border.color: "black"
		Text {
	    	id: conv
	    	objectName: "conv"
	    	text: ""
	    	wrapMode: "WordWrap"
	    	anchors.fill: parent
		}
     }
}
