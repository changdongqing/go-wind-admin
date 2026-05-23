import {Card, Row, Col} from 'antd';
import {
    UserOutlined,
    EyeOutlined,
    DownloadOutlined,
    ClockCircleOutlined
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import {usePreferences} from '@/core/preferences';

const Dashboard = () => {
    const {isDark} = usePreferences();

    // 统计卡片数据
    const statsData = [
        {
            title: '用户量',
            value: '2,000',
            total: '总用户量',
            totalValue: '120,000',
            icon: <UserOutlined style={{fontSize: 24, color: '#1677ff'}}/>,
        },
        {
            title: '访问量',
            value: '20,000',
            total: '总访问量',
            totalValue: '500,000',
            icon: <EyeOutlined style={{fontSize: 24, color: '#ff7a45'}}/>,
        },
        {
            title: '下载量',
            value: '8,000',
            total: '总下载量',
            totalValue: '120,000',
            icon: <DownloadOutlined style={{fontSize: 24, color: '#faad14'}}/>,
        },
        {
            title: '使用量',
            value: '5,000',
            total: '总使用量',
            totalValue: '50,000',
            icon: <ClockCircleOutlined style={{fontSize: 24, color: '#52c41a'}}/>,
        },
    ];

    // 访问趋势图配置
    const trendOption = {
        tooltip: {
            trigger: 'axis',
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true,
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: ['6:00', '7:00', '8:00', '9:00', '10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00', '20:00', '21:00', '22:00', '23:00'],
            axisLine: {
                lineStyle: {
                    color: isDark ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.1)',
                },
            },
        },
        yAxis: {
            type: 'value',
            axisLine: {
                lineStyle: {
                    color: isDark ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.1)',
                },
            },
            splitLine: {
                lineStyle: {
                    color: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.05)',
                },
            },
        },
        series: [
            {
                name: '访问量',
                type: 'line',
                smooth: true,
                areaStyle: {
                    opacity: 0.3,
                    color: '#1677ff',
                },
                lineStyle: {
                    width: 2,
                    color: '#1677ff',
                },
                itemStyle: {
                    color: '#1677ff',
                },
                data: [0, 5000, 15000, 25000, 40000, 55000, 65000, 45000, 20000, 35000, 50000, 70000, 45000, 25000, 15000, 8000, 4000, 2000],
            },
            {
                name: '趋势',
                type: 'line',
                smooth: true,
                areaStyle: {
                    opacity: 0.3,
                    color: '#52c41a',
                },
                lineStyle: {
                    width: 2,
                    color: '#52c41a',
                },
                itemStyle: {
                    color: '#52c41a',
                },
                data: [0, 1000, 3000, 8000, 15000, 20000, 22000, 15000, 8000, 12000, 18000, 23000, 15000, 8000, 4000, 2000, 1000, 500],
            },
        ],
    };

    return (
        <div>
            {/* 统计卡片 */}
            <Row gutter={[16, 16]}>
                {statsData.map((stat, index) => (
                    <Col xs={24} sm={12} lg={6} key={index}>
                        <Card
                            style={{
                                height: '100%',
                            }}
                        >
                            <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start'}}>
                                <div>
                                    <div style={{fontSize: 14, color: isDark ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)', marginBottom: 8}}>
                                        {stat.title}
                                    </div>
                                    <div style={{fontSize: 24, fontWeight: 600, marginBottom: 16}}>
                                        {stat.value}
                                    </div>
                                    <div style={{display: 'flex', justifyContent: 'space-between', fontSize: 12}}>
                                        <span style={{color: isDark ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.45)'}}>
                                            {stat.total}
                                        </span>
                                        <span style={{color: isDark ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)'}}>
                                            {stat.totalValue}
                                        </span>
                                    </div>
                                </div>
                                <div>{stat.icon}</div>
                            </div>
                        </Card>
                    </Col>
                ))}
            </Row>

            {/* 访问趋势图 */}
            <Card style={{marginTop: 16}}>
                <ReactECharts option={trendOption} style={{height: 300}}/>
            </Card>
        </div>
    );
};

export default Dashboard;

