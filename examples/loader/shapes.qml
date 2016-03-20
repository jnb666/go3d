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

    toolBar:ToolBar {
        RowLayout {
            ToolButton {
                text: "spin"; checkable: true
                onClicked: anim.running = checked
            }
            ToolButton {
                text: "scenery"; checkable: true
                onClicked: shapes.setScenery(checked)
            }            
            ComboBox {
                model: ["cube", "teapot", "shuttle", "bunny", "dragon"]
                onCurrentIndexChanged: shapes.setModel(currentText)
            }
        }
    }
    Shapes {
        id: shapes
        anchors.fill: parent
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

