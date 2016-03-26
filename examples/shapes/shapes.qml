import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import QtQuick.Dialogs 1.0
import GoExtensions 1.0

ApplicationWindow {
    id: root
    title: "shapes"
    x: 100; y: 30; minimumWidth: 640; minimumHeight: 640
    color: "#404040"

    ColumnLayout {
        RowLayout {
            id: menu
            spacing: 20
            anchors.margins: 5
            anchors.top: parent.top
            anchors.left: parent.left
            Button {
                text: "spin"; checkable: true
                onClicked: anim.running = checked
            }
            ComboBox {
                model: ["cube", "prism", "pyramid", "point", "plane", "circle", "cylinder", "cone", "icosohedron", "sphere"]
                onCurrentIndexChanged: shapes.setShape(currentText)
            }
            ComboBox {
                model: ["plastic", "wood", "rough", "marble", "metallic", "glass", "earth", "emissive", "diffuse", "unshaded", "texture2d", "texturecube"]
                onCurrentIndexChanged: shapes.setMaterial(currentText)
            }
            Button {
                text: "choose colour" 
                onClicked: {
                    colorDialog.color = shapes.getColor()
                    colorDialog.open()
                }
            }
            Button {
                text: "scenery"; checkable: true
                onClicked: shapes.setScenery(checked)
            }
        }
        Shapes {
            id: shapes
            anchors.left: parent.left
            anchors.top: menu.bottom
            anchors.margins: 5
            width: root.width-10
            height: root.height-menu.height-10
            Timer {
                id: anim
                interval: 20; running: false; repeat: true
                onTriggered: shapes.spin()
            }
            MouseArea {
                anchors.fill: parent
                onWheel: shapes.zoom(wheel.angleDelta.y)
                acceptedButtons: Qt.LeftButton | Qt.RightButton
                onPressed: shapes.mouse("start", mouse.x, mouse.y, mouse.button)
                onPositionChanged: shapes.mouse("move", mouse.x, mouse.y, mouse.button)
                onReleased: shapes.mouse("end", 0, 0, mouse.button)
            }
        }
    }
    ColorDialog {
        id: colorDialog
        title: "Please choose a color"
        showAlphaChannel: true
        onAccepted: shapes.setColor(colorDialog.currentColor)
    }
}

