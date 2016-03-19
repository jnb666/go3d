import QtQuick 2.2
import QtQuick.Controls 1.1
import GoExtensions 1.0

ApplicationWindow {
    id: root
    title: "triangle"
    minimumWidth: 500; minimumHeight: 500
    color: "black"

    Triangle {
        id: triangle
        anchors.fill: parent
        Timer {
            interval: 20; running: true; repeat: true
            onTriggered: triangle.rotate()
        }
    }
}