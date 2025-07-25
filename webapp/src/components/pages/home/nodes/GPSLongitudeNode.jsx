// SPDX-FileCopyrightText: © 2023 OneEyeFPV oneeyefpv@gmail.com
// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-License-Identifier: FS-0.9-or-later

import React from 'react';

import {GenericTelemetryNode} from "./GenericTelemetryNode";
import GPSLongitudeTelemetryIcon from "../../../icons/GPSLongitudeTelemetryIcon";
import {GPSData} from "../../../../pbwrap";

function GPSLongitudeNode(node) {
    return (<GenericTelemetryNode
        valueProps={{
            isValidTelemetryData: data => data instanceof GPSData,
            getTelemetryValue: data => data.getLongitude()
        }}
        node={node}/>);

}

GPSLongitudeNode.type = "gps_longitude";
GPSLongitudeNode.menuIcon = <GPSLongitudeTelemetryIcon />;


export default GPSLongitudeNode;
