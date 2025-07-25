// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

import React from 'react';

import {GenericTelemetryNode} from "./GenericTelemetryNode";
import GPSLatitudeTelemetryIcon from "../../../icons/GPSLatitudeTelemetryIcon";
import {GPSData} from "../../../../pbwrap";

function GPSLatitudeNode(node) {
    return (<GenericTelemetryNode
        valueProps={{
            isValidTelemetryData: data => data instanceof GPSData,
            getTelemetryValue: data => data.getLatitude()
        }}
        node={node}/>);

}

GPSLatitudeNode.type = "gps_latitude";
GPSLatitudeNode.menuIcon = <GPSLatitudeTelemetryIcon />;


export default GPSLatitudeNode;
