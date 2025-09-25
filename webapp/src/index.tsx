// Copyright (c) 2015-present Mattermost, Inc.
// All Rights Reserved. See LICENSE.txt for license information.

import React from 'react';

import manifest from '@/manifest';
import type {PluginRegistry} from '@/types/mattermost-webapp';

const TelemostPost: React.FC<{post: any}> = ({post}) => {
    const joinURL = post.props?.joinURL;
    const meetingID = post.props?.meetingID;
    const title = post.props?.title || 'Telemost Meeting';
    const pretext = post.props?.pretext || 'I have started a meeting';

    // Auto-open functionality removed to prevent unwanted redirects

    if (!joinURL) {
        return null;
    }

    return (
        <div className="attachment attachment--pretext">
            {/* Pretext ( "I have started a meeting") */}
            <div className="attachment__thumb-pretext">
                <p>{pretext}</p>
            </div>

            {/* Main content */}
            <div className="attachment__content">
                <div className="clearfix attachment__container">
                    {/* Title */}
                    <h5 className="mt-1" style={{fontWeight: 600}}>
                        <p>{title}</p>
                    </h5>

                    {/* Meeting ID */}
                    <span>
                        Meeting ID :{' '}
                        <a
                            rel="noopener noreferrer"
                            target="_blank"
                            href={joinURL}
                        >
                            {meetingID}
                        </a>
                    </span>

                    {/* Join button */}
                    <div>
                        <div style={{overflow: 'auto hidden', paddingRight: '5px', width: '100%'}}>
                            <a
                                className="btn btn-primary"
                                rel="noopener noreferrer"
                                target="_blank"
                                href={joinURL}
                                style={{
                                    fontFamily: '"Open Sans", sans-serif',
                                    fontSize: '12px',
                                    fontWeight: 'bold',
                                    letterSpacing: '1px',
                                    lineHeight: '19px',
                                    marginTop: '12px',
                                    marginRight: '12px',
                                    borderRadius: '4px',
                                    color: '#fff',
                                    backgroundColor: '#e56a52',
                                    padding: '6px 12px',
                                    display: 'inline-flex',
                                    alignItems: 'center',
                                    textDecoration: 'none'
                                }}
                            >
                                {/* Video Icon */}
                                <i style={{paddingRight: '8px', display: 'flex'}}>
                                    <svg
                                        width="19px"
                                        height="100%"
                                        viewBox="0 0 19 10"
                                        xmlns="http://www.w3.org/2000/svg"
                                        fill="white"
                                    >
                                        <path d="M1,0 L10,0 C12.2,0 14,1.8 14,4 L14,9 C14,9.6 13.6,10 13,10 L4,10 C1.8,10 0,8.2 0,6 L0,1 C0,0.4 0.4,0 1,0 Z"></path>
                                        <path d="M15.4,2.9 L17.4,1.2 C17.8,0.9 18.4,0.9 18.8,1.4 C18.9,1.5 19,1.8 19,2 V9 C19,9.6 18.6,10 18,10 C17.8,10 17.5,9.9 17.4,9.8 L15.4,8.1 C15.1,7.9 15,7.7 15,7.4 V3.6 C15,3.3 15.1,3.1 15.4,2.9 Z"></path>
                                    </svg>
                                </i>
                                JOIN MEETING
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

// Connection status component
const TelemostConnection: React.FC<{post: any}> = ({post}) => {
    const userId = post.props?.userId;
    const username = post.props?.username;
    const status = post.props?.status;

    if (!userId || !username || !status) {
        return null;
    }

    const isConnected = status === 'connected';
    const statusColor = isConnected ? '#28a745' : '#dc3545';
    const statusText = isConnected ? 'Connected' : 'Disconnected';
    const statusIcon = isConnected ? '🔗' : '❌';

    return (
        <div className="attachment attachment--pretext">
            <div className="attachment__content" style={{borderLeftColor: statusColor}}>
                <div className="clearfix attachment__container">
                    <h5 className="mt-1" style={{fontWeight: 600, margin: '0 0 8px 0'}}>
                        <span style={{marginRight: '8px'}}>{statusIcon}</span>
                        Telemost Connection Status
                    </h5>
                    <p style={{margin: '0 0 8px 0'}}>
                        <strong>{username}</strong> is now <strong style={{color: statusColor}}>{statusText}</strong>
                    </p>
                    <div style={{fontSize: '12px', color: '#666'}}>
                        {isConnected 
                            ? 'You are now connected to Telemost. You can start meetings or join existing ones.'
                            : 'You have disconnected from Telemost. Use /telemost connect to reconnect.'
                        }
                    </div>
                </div>
            </div>
        </div>
    );
};

// Telemost Icon Component
const TelemostIcon: React.FC<{useSVG?: boolean}> = ({useSVG = true}) => {
    if (useSVG) {
        return (
            <span 
                aria-label="Telemost video icon"
                style={{position: 'relative', top: '-1px'}}
                dangerouslySetInnerHTML={{
                    __html: `<svg width='20' height='20' viewBox='0 0 24 24' fill='none' xmlns='http://www.w3.org/2000/svg' style='vertical-align: middle;'>
                        <circle cx='12' cy='12' r='12' fill='#e56a52'/>
                        <path d='M5.5 9.5C5.5 8.7 6.2 8 7 8H11C12.4 8 13.5 9.1 13.5 10.5V13.5C13.5 14.3 12.8 15 12 15H8C6.6 15 5.5 13.9 5.5 12.5V9.5Z' fill='white'/>
                        <path fill-rule='evenodd' clip-rule='evenodd' d='M17.8 15.2L16.2 14.2C15.9 14 15.8 13.7 15.8 13.4V10.6C15.8 10.3 15.9 10 16.2 9.8L17.8 8.8C18.2 8.5 18.7 8.5 19 8.8C19.2 8.9 19.3 9.1 19.3 9.3V14.7C19.3 15.1 19 15.4 18.6 15.4C18.4 15.4 18.2 15.3 18 15.2L17.8 15.2Z' fill='white'/>
                    </svg>`
                }}
            />
        );
    }
    
    return (
        <span aria-label="Telemost video icon">
            <i className="icon icon-brand-telemost"></i>
        </span>
    );
};

export default class Plugin {
    public async initialize(registry: PluginRegistry) {
        console.log('Telemost plugin initializing...');
        
        // Register meeting component
        registry.registerPostTypeComponent('custom_telemost_meeting', TelemostPost);
        console.log('Telemost plugin registered custom_telemost_meeting component');
        
        // Register connection component
        registry.registerPostTypeComponent('custom_telemost_connection', TelemostConnection);
        console.log('Telemost plugin registered custom_telemost_connection component');
        
        // Register channel header button
        registry.registerChannelHeaderButtonAction(
            <TelemostIcon useSVG={true} />,
            () => {
                // Start a Telemost meeting in this channel
                window.open(`/plugins/${manifest.id}/start`, '_blank');
            },
            'Start Telemost Meeting',
            'Start Telemost Meeting',
        );
        console.log('Telemost plugin registered channel header button');
        
        // Register app bar component
        if (registry.registerAppBarComponent) {
            const appBarIconUrl = `/plugins/${manifest.id}/public/app-bar-icon.png`;
            registry.registerAppBarComponent(
                appBarIconUrl,
                async () => {
                    // Start a Telemost meeting
                    window.open(`/plugins/${manifest.id}/start`, '_blank');
                },
                'Start Telemost Meeting',
                'all',
            );
            console.log('Telemost plugin registered app bar component');
        }
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
