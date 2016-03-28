import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import QtQuick.Dialogs 1.0
import GoExtensions 1.0

ApplicationWindow {
    id: root
    title: "model"
    x: 100; y: 30; minimumWidth: 1000; minimumHeight: 800

    toolBar:ToolBar {
        width: root.width
        RowLayout {
            anchors.fill: parent
            Button {
                text: "spin"; checkable: true
                onClicked: spin.running = checked
            }
            Button {
                text: "scenery"; checkable: true; checked: true
                onClicked: model.setScenery(checked)
            }
            Button {
                text: "bump maps"; checkable: true; checked: false
                onClicked: model.enableBump(checked)
            }
            Button {
                text: "reset" 
                onClicked: { model.reset(); move.running = false; moveButton.checked = false }
            }
            Button {
                id: moveButton; text: "move"; checkable: true
                onClicked: move.running = checked
            }
            ComboBox {
                model: ["arc ball camera", "first person camera"]
                onCurrentIndexChanged: model.setCamera(currentIndex)
            }
            ComboBox {
                model: ["cube", "teapot", "shuttle", "bunny", "dragon", "sponza", "sibenik"]
                onCurrentIndexChanged: model.setModel(currentText)
            }
        }
    }
    Model {
        id: model; anchors.fill: parent
        Timer {
            id: move; interval: 20; running: false; repeat: true
            onTriggered: model.move(0.25)
        }
        Timer {
            id: spin; interval: 20; running: false; repeat: true
            onTriggered: model.spin()
        }
        focus: true
        Keys.onPressed: {
            if (event.key == Qt.Key_Left) {
                model.rotate("keys", -2, 0, 0)
            }
            if (event.key == Qt.Key_Right) {
                model.rotate("keys", 2, 0, 0)
            }
            if (event.key == Qt.Key_Up) {
                model.rotate("keys", 0, -2, 0)
            }
            if (event.key == Qt.Key_Down) {
                model.rotate("keys", 0, 2, 0)
            }
            if (event.key == Qt.Key_Space) {
                model.move(2)
            }
        }
        MouseArea {
            anchors.fill: parent
            onWheel: model.move(wheel.angleDelta.y > 0 ? 1 : -1)
            acceptedButtons: Qt.LeftButton | Qt.RightButton
            onPressed: model.rotate("start", mouse.x, mouse.y, mouse.button)
            onPositionChanged: model.rotate("move", mouse.x, mouse.y, mouse.button)
            onReleased: model.rotate("end", 0, 0, mouse.button)
        }
    }
}

