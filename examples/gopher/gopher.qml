import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import GoExtensions 1.0

ApplicationWindow {
    id: root
    title: "gopher"
    minimumWidth: 500; minimumHeight: 500
    color: "#404040"

    ColumnLayout {
        RowLayout {
            id: menu
            spacing: 20
            anchors.margins: 5
            anchors.left: parent.left
            Button {
                text: "spin"; checkable: true
                onClicked: {
                    anim.running = checked
                    slider.updateValue()
                }
            }
            Slider {
                id: slider
                value: 50
                minimumValue: -200
                maximumValue: 200
                Component.onCompleted: {
                    slider.onValueChanged.connect(updateValue)
                }
                function updateValue() {               
                    anim.interval = Math.abs(1000.0/value)
                    gopher.setStep(value > 0 ? 1 : -1)
                }
            }
        }
        GopherCube {
            id: gopher
            anchors.left: parent.left
            anchors.top: menu.bottom
            anchors.margins: 5
            width: root.width-10
            height: root.height-menu.height-10        
            Timer {
                id: anim
                interval: 20; running: false; repeat: true
                onTriggered: gopher.rotate()
            }
        }
    }
}


                