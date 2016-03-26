import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import GoExtensions 1.0

ApplicationWindow {
    id: root
    title: "lighting"
    x: 100; y: 30; minimumWidth: 640; minimumHeight: 640
    color: "black"

    toolBar:ToolBar {
        RowLayout {
            ToolButton {
                text: "spin model"; checkable: true; onClicked: anim.running = checked
            }
            ToolButton {
                text: "reset view"; onClicked: model.reset()
            }
            ComboBox {
                model: ["diffuse", "specular", "normal"]
                onCurrentIndexChanged: model.setLighting(currentText)
            }
            ComboBox {
                model: ["brick", "shield"]
                onCurrentIndexChanged: model.setTexture(currentText)
            }
        }
    }
    Model {
        id: model
        anchors.fill: parent
        Timer {
            id: anim
            interval: 20; running: false; repeat: true
            onTriggered: model.spin()
        }
        MouseArea {
            anchors.fill: parent
            onWheel: model.zoom(wheel.angleDelta.y)
            acceptedButtons: Qt.LeftButton | Qt.RightButton
            onPressed: model.mouse("start", mouse.x, mouse.y, mouse.button)
            onPositionChanged: model.mouse("move", mouse.x, mouse.y, mouse.button)
            onReleased: model.mouse("end", 0, 0, mouse.button)
        }
    }
}