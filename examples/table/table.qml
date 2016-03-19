import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import GoExtensions 1.0

ApplicationWindow {
    id: root
    title: "table scene"
    x: 100; y: 30; minimumWidth: 700; minimumHeight: 700
    color: "#404040"

    toolBar:ToolBar {
        RowLayout {
            ToolButton {
                text: "light1"; checkable: true; checked: true
                onClicked: scene.showLight(0, checked)
            }
            ToolButton {
                text: "light2"; checkable: true; checked: true
                onClicked: scene.showLight(1, checked)
            }
            ToolButton {
                text: "table"; checkable: true; checked: true
                onClicked: scene.showTable(checked)
            }
        }
    }

    Scene {
        id: scene
        anchors.fill: parent
        focus: true
        MouseArea {
            anchors.fill: parent
            onWheel: scene.zoom(wheel.angleDelta.y)         
            onPressed: scene.mouse("start", mouse.x, mouse.y)
            onPositionChanged: scene.mouse("move", mouse.x, mouse.y)
            onReleased: scene.mouse("end", 0, 0)
        }
    }
}

